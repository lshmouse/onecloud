// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package identity

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	api "yunion.io/x/onecloud/pkg/apis/identity"
	"yunion.io/x/onecloud/pkg/cloudcommon/consts"
	"yunion.io/x/onecloud/pkg/cloudcommon/policy"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/auth"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
	"yunion.io/x/onecloud/pkg/util/fileutils2"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/onecloud/pkg/util/shellutils"
)

func createPolicy(s *mcclient.ClientSession, name string, policy string, domain string, enabled bool, disabled bool, desc string, scope string, isSystem *bool) error {
	params := jsonutils.NewDict()
	params.Add(jsonutils.NewString(name), "type")
	params.Add(jsonutils.NewString(policy), "policy")
	if len(domain) > 0 {
		params.Add(jsonutils.NewString(domain), "project_domain_id")
	}
	if enabled {
		params.Add(jsonutils.JSONTrue, "enabled")
	} else if disabled {
		params.Add(jsonutils.JSONFalse, "enabled")
	}
	if len(desc) > 0 {
		params.Add(jsonutils.NewString(desc), "description")
	}
	if len(scope) > 0 {
		params.Add(jsonutils.NewString(scope), "scope")
	}
	if isSystem != nil {
		if *isSystem {
			params.Add(jsonutils.JSONTrue, "is_system")
		} else {
			params.Add(jsonutils.JSONFalse, "is_system")
		}
	}

	result, err := modules.Policies.Create(s, params)
	if err != nil {
		return err
	}

	printObject(result)

	return nil

}

func init() {
	type PolicyListOptions struct {
		options.BaseListOptions
		Type          string `help:"filter by type"`
		IsSystem      *bool  `help:"filter by is_system" negative:"is_no_system"`
		Format        string `help:"policy format, default to yaml" default:"yaml" choices:"yaml|json"`
		OrderByDomain string `help:"order by domain name" choices:"asc|desc"`
	}
	R(&PolicyListOptions{}, "policy-list", "List all policies", func(s *mcclient.ClientSession, args *PolicyListOptions) error {
		params, err := options.ListStructToParams(args)
		if err != nil {
			return err
		}
		result, err := modules.Policies.List(s, params)
		if err != nil {
			return err
		}
		printList(result, modules.Policies.GetColumns(s))
		return nil
	})

	type PolicyCreateOptions struct {
		Domain   string `help:"domain of the policy"`
		NAME     string `help:"name of the policy"`
		FILE     string `help:"path to policy file"`
		Enabled  bool   `help:"create policy enabled"`
		Disabled bool   `help:"create policy disabled"`
		Desc     string `help:"policy description"`
		Scope    string `help:"scope of policy"`
		IsSystem *bool  `help:"create system policy" negative:"no-system"`
	}
	R(&PolicyCreateOptions{}, "policy-create", "Create a new policy", func(s *mcclient.ClientSession, args *PolicyCreateOptions) error {
		policyBytes, err := ioutil.ReadFile(args.FILE)
		if err != nil {
			return err
		}

		return createPolicy(s, args.NAME, string(policyBytes), args.Domain, args.Enabled, args.Disabled, args.Desc, args.Scope, args.IsSystem)
	})

	type PolicyCloneOptions struct {
		Domain string `help:"domain of the policy"`
		OLD    string `help:"name or id of the old policy"`
		NAME   string `help:"name of the policy"`
	}
	R(&PolicyCloneOptions{}, "policy-clone", "Clone a policy", func(s *mcclient.ClientSession, args *PolicyCloneOptions) error {
		old, err := modules.Policies.Get(s, args.OLD, nil)
		if err != nil {
			return err
		}
		policy, _ := old.GetString("policy")
		enabled, _ := old.Bool("enabled")
		desc, _ := old.GetString("description")
		scope, _ := old.GetString("scope")
		isSystem, _ := old.Bool("is_system")
		return createPolicy(s, args.NAME, policy, args.Domain, enabled, !enabled, desc, scope, &isSystem)
	})

	type PolicyPatchOptions struct {
		ID       string `help:"ID of policy"`
		File     string `help:"path to policy file"`
		Type     string `help:"policy type"`
		Enabled  bool   `help:"update policy enabled"`
		Disabled bool   `help:"update policy disabled"`
		Desc     string `help:"Description"`
		IsSystem bool   `help:"is_system"`

		IsNotSystem bool `help:"negative is_system"`
	}
	updateFunc := func(s *mcclient.ClientSession, args *PolicyPatchOptions) error {
		policyId, err := modules.Policies.GetId(s, args.ID, nil)
		if err != nil {
			return err
		}
		params := jsonutils.NewDict()
		if len(args.Type) > 0 {
			params.Add(jsonutils.NewString(args.Type), "type")
		}
		if len(args.File) > 0 {
			policyBytes, err := ioutil.ReadFile(args.File)
			if err != nil {
				return err
			}
			params.Add(jsonutils.NewString(string(policyBytes)), "policy")
		}
		if args.Enabled {
			params.Add(jsonutils.JSONTrue, "enabled")
		} else if args.Disabled {
			params.Add(jsonutils.JSONFalse, "enabled")
		}
		if len(args.Desc) > 0 {
			params.Add(jsonutils.NewString(args.Desc), "description")
		}
		if args.IsSystem {
			params.Add(jsonutils.JSONTrue, "is_system")
		}
		if args.IsNotSystem {
			params.Add(jsonutils.JSONFalse, "is_system")
		}
		result, err := modules.Policies.Update(s, policyId, params)
		if err != nil {
			return err
		}

		printObject(result)

		return nil
	}
	R(&PolicyPatchOptions{}, "policy-patch", "Patch policy", updateFunc)
	R(&PolicyPatchOptions{}, "policy-update", "Update policy", updateFunc)

	type PolicyPublicOptions struct {
		ID            string   `help:"ID of policy to update" json:"-"`
		Scope         string   `help:"sharing scope" choices:"system|domain"`
		SharedDomains []string `help:"share to domains"`
	}
	R(&PolicyPublicOptions{}, "policy-public", "Mark a policy public", func(s *mcclient.ClientSession, args *PolicyPublicOptions) error {
		params := jsonutils.Marshal(args)
		result, err := modules.Policies.PerformAction(s, args.ID, "public", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type PolicyPrivateOptions struct {
		ID string `help:"ID of policy to update" json:"-"`
	}
	R(&PolicyPrivateOptions{}, "policy-private", "Mark a policy private", func(s *mcclient.ClientSession, args *PolicyPrivateOptions) error {
		result, err := modules.Policies.PerformAction(s, args.ID, "private", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type PolicyDeleteOptions struct {
		ID string `help:"ID of policy"`
	}
	R(&PolicyDeleteOptions{}, "policy-delete", "Delete policy", func(s *mcclient.ClientSession, args *PolicyDeleteOptions) error {
		policyId, err := modules.Policies.GetId(s, args.ID, nil)
		if err != nil {
			return err
		}
		result, err := modules.Policies.Delete(s, policyId, nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type PolicyShowOptions struct {
		ID     string `help:"ID of policy"`
		Format string `help:"policy format, default yaml" default:"yaml" choices:"yaml|json"`
		Save   string `help:"save policy data into a file"`
	}
	R(&PolicyShowOptions{}, "policy-show", "Show policy", func(s *mcclient.ClientSession, args *PolicyShowOptions) error {
		query := jsonutils.NewDict()
		query.Add(jsonutils.NewString(args.Format), "format")
		result, err := modules.Policies.Get(s, args.ID, query)
		if err != nil {
			return err
		}
		printObject(result)
		if len(args.Save) > 0 {
			if args.Format == "yaml" {
				yaml, _ := result.GetString("policy")
				fileutils2.FilePutContents(args.Save, yaml, false)
			} else {
				p, _ := result.Get("policy")
				fileutils2.FilePutContents(args.Save, p.PrettyString(), false)
			}
		}
		return nil
	})

	type PolicyEditOptions struct {
		ID string `help:"ID of policy"`
	}
	R(&PolicyEditOptions{}, "policy-edit", "Edit and update policy", func(s *mcclient.ClientSession, args *PolicyEditOptions) error {
		query := jsonutils.NewDict()
		query.Add(jsonutils.NewString("yaml"), "format")
		result, err := modules.Policies.Get(s, args.ID, query)
		if err != nil {
			return err
		}
		policyId, err := result.GetString("id")
		if err != nil {
			return err
		}
		yaml, err := result.GetString("policy")
		if err != nil {
			return err
		}

		yaml, err = shellutils.Edit(yaml)
		if err != nil {
			return err
		}

		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(yaml), "policy")

		result, err = modules.Policies.Patch(s, policyId, params)
		if err != nil {
			return err
		}

		printObject(result)
		return nil
	})

	type PolicyBindRoleOptions struct {
		POLICY string `json:"-" help:"policy Id or name"`
		api.PolicyBindRoleInput
	}
	R(&PolicyBindRoleOptions{}, "policy-bind-role", "Policy bind role", func(s *mcclient.ClientSession, args *PolicyBindRoleOptions) error {
		params := jsonutils.Marshal(args)
		log.Debugf("params: %s", params)
		result, err := modules.Policies.PerformAction(s, args.POLICY, "bind-role", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type PolicyAdminCapableOptions struct {
		User          string   `help:"For user"`
		UserDomain    string   `help:"Domain for user"`
		Project       string   `help:"Role assignments for project"`
		ProjectDomain string   `help:"Domain for project"`
		Role          []string `help:"Role name list"`
		RoleId        []string `help:"Role Id list"`
	}
	R(&PolicyAdminCapableOptions{}, "policy-admin-capable", "Check admin capable", func(s *mcclient.ClientSession, args *PolicyAdminCapableOptions) error {
		auth.InitFromClientSession(s)
		policy.EnableGlobalRbac(15*time.Second, false)

		var token mcclient.TokenCredential
		if len(args.User) > 0 {
			token = &mcclient.SSimpleToken{
				Domain:        args.UserDomain,
				User:          args.User,
				Project:       args.Project,
				ProjectDomain: args.ProjectDomain,
				Roles:         strings.Join(args.Role, ","),
				RoleIds:       strings.Join(args.RoleId, ","),
			}
		} else {
			token = s.GetToken()
		}

		fmt.Println("Token", token)

		for _, scope := range []rbacutils.TRbacScope{
			rbacutils.ScopeSystem,
			rbacutils.ScopeDomain,
			rbacutils.ScopeProject,
			rbacutils.ScopeUser,
			rbacutils.ScopeNone,
		} {
			fmt.Printf("%s: ", scope)
			capable := policy.PolicyManager.IsScopeCapable(token, scope)
			fmt.Println(capable)
		}

		return nil
	})

	type PolicyExplainOptions struct {
		User          string   `help:"For user"`
		UserDomain    string   `help:"Domain for user"`
		Project       string   `help:"Role assignments for project"`
		Role          []string `help:"Roles"`
		RoleId        []string `help:"Role Id list"`
		Request       []string `help:"explain request, in format of key:scope:service:resource:action:extra"`
		Name          string   `help:"policy name"`
		Debug         bool     `help:"enable RBAC debug"`
		ProjectDomain string   `help:"Domain name"`
		Ip            string   `help:"login IP"`
	}
	R(&PolicyExplainOptions{}, "policy-explain", "Explain policy result", func(s *mcclient.ClientSession, args *PolicyExplainOptions) error {
		err := log.SetLogLevelByString(log.Logger(), "debug")
		if err != nil {
			log.Fatalf("Set log level %q: %v", "debug", err)
		}
		if args.Debug {
			rbacutils.ShowMatchRuleDebug = true
		}
		auth.InitFromClientSession(s)
		policy.EnableGlobalRbac(15*time.Second, false)
		if args.Debug {
			consts.EnableRbacDebug()
		}

		req := jsonutils.NewDict()
		for i := 0; i < len(args.Request); i += 1 {
			parts := strings.Split(args.Request[i], ":")
			if len(parts) < 3 {
				return fmt.Errorf("invalid request, should be in the form of key:[system|domain|project]:service[:resource:action:extra]")
			}
			key := parts[0]
			data := make([]jsonutils.JSONObject, 0)
			for i := 1; i < len(parts); i += 1 {
				data = append(data, jsonutils.NewString(parts[i]))
			}
			req.Add(jsonutils.NewArray(data...), key)
		}
		fmt.Println("Request:", req.String())

		var token mcclient.TokenCredential
		if len(args.User) > 0 {
			usrParams := jsonutils.NewDict()
			if len(args.UserDomain) > 0 {
				usrDom, err := modules.Domains.Get(s, args.UserDomain, nil)
				if err != nil {
					return fmt.Errorf("search user domain %s fail %s", args.UserDomain, err)
				}
				usrDomId, _ := usrDom.Get("id")
				usrParams.Add(usrDomId, "domain_id")
			}
			usr, err := modules.UsersV3.Get(s, args.User, usrParams)
			if err != nil {
				return fmt.Errorf("search user %s fail %s", args.User, err)
			}
			usrId, _ := usr.GetString("id")
			usrName, _ := usr.GetString("name")
			usrDomId, _ := usr.GetString("domain_id")
			usrDom, _ := usr.GetString("domain")

			projParams := jsonutils.NewDict()
			if len(args.ProjectDomain) > 0 {
				projDom, err := modules.Domains.Get(s, args.ProjectDomain, nil)
				if err != nil {
					return fmt.Errorf("search project domain %s fail %s", args.ProjectDomain, err)
				}
				projDomId, _ := projDom.Get("id")
				projParams.Add(projDomId, "domain_id")
			}
			proj, err := modules.Projects.Get(s, args.Project, projParams)
			if err != nil {
				return fmt.Errorf("search project %s fail %s", args.Project, err)
			}
			projId, _ := proj.GetString("id")
			projName, _ := proj.GetString("name")
			projDom, _ := proj.GetString("domain")
			projDomId, _ := proj.GetString("domain_id")
			token = &mcclient.SSimpleToken{
				Domain:          usrDom,
				DomainId:        usrDomId,
				User:            usrName,
				UserId:          usrId,
				Project:         projName,
				ProjectId:       projId,
				ProjectDomain:   projDom,
				ProjectDomainId: projDomId,
				Roles:           strings.Join(args.Role, ","),
				RoleIds:         strings.Join(args.RoleId, ","),
				Context: mcclient.SAuthContext{
					Ip: args.Ip,
				},
				Token: "faketoken",
			}
		} else {
			token = s.GetToken()
		}

		result, err := policy.ExplainRpc(context.Background(), token, req, args.Name)
		if err != nil {
			return err
		}
		printObject(result)

		/*for _, r := range args.Role {
			fmt.Println("role", r, "matched policies:", policy.PolicyManager.RoleMatchPolicies(r))
		}

		fmt.Println("userCred:", token)
		for _, scope := range []rbacutils.TRbacScope{
			rbacutils.ScopeSystem,
			rbacutils.ScopeDomain,
			rbacutils.ScopeProject,
			rbacutils.ScopeUser,
			rbacutils.ScopeNone,
		} {
			m := policy.PolicyManager.MatchedPolicyNames(scope, token)
			fmt.Println("matched", scope, "policies:", m)
		}

		fmt.Println("all_policies", policy.PolicyManager.AllPolicies())*/
		return nil
	})
}
