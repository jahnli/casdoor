import * as React from "react";
import i18next from "i18next";
import {CalendarDays} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {Link, useNavigate, useParams, useSearchParams} from "react-router-dom";
import {UserCell} from "@/components/user/UserCell";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {XlsxImport} from "@/components/crud/XlsxImport";
import {boolColumn, dateColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";
import {newUser} from "@/pages/defaults";

/**
 * Also embedded by the group tree page, which scopes the list to the selected
 * group through props instead of the route.
 */
export default function UserListPage({
  organizationName: organizationNameProp,
  groupName: groupNameProp,
  formItems,
}: {organizationName?: string; groupName?: string; formItems?: any[]} = {}) {
  const {account} = useAccount();
  const params = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const embedded = groupNameProp !== undefined;
  const groupName = groupNameProp ?? searchParams.get("groupName") ?? "";
  const scopedOrganizationName = organizationNameProp ?? params.organizationName;
  const organizationName = useRequestOrganization(scopedOrganizationName);

  // antd stops you removing or deleting yourself, or the built-in admin
  const isProtected = (record: any) =>
    (record.owner === account?.owner && record.name === account?.name) ||
    (record.owner === "built-in" && record.name === "admin");

  const [organization, setOrganization] = React.useState<any>({});

  React.useEffect(() => {
    if (!organizationName) {
      return;
    }
    OrganizationBackend.getOrganization("admin", organizationName).then((res: any) => {
      if (res.status === "ok") {
        setOrganization(res.data ?? {});
      }
    });
  }, [organizationName]);

  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) && !scopedOrganizationName : false;

  const columns: ColumnDef<any>[] = [
    {
      dataIndex: "displayName",
      title: i18next.t("general:User"),
      width: 200,
      fixed: "left",
      sortable: true,
      searchable: true,
      searchPlaceholder: i18next.t("general:Search name or username"),
      render: (_value, record) => (
        <div className="space-y-1">
          <UserCell user={record} />
          {record.pendingApproval ? <Badge variant="outline">{i18next.t("user:Pending approval")}</Badge> : null}
        </div>
      ),
    },
    textColumn({dataIndex: "affiliation", title: i18next.t("user:Affiliation"), width: 140, searchable: true}),
    {
      dataIndex: "department",
      title: i18next.t("user:Employment overview"),
      width: 260,
      render: (_value, record) => {
        const properties = record.properties ?? {};
        const department = String(properties.department ?? "").trim();
        const jobLevel = String(properties.job_level ?? "").trim();
        const workDate = String(properties.job_date ?? "").trim();
        return (
          <div className="max-w-[280px] space-y-1.5 overflow-hidden font-normal">
            {department || (!jobLevel && !workDate) ? (
              department ? (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div className="truncate text-sm" tabIndex={0}>{department}</div>
                  </TooltipTrigger>
                  <TooltipContent className="whitespace-pre-wrap break-words" side="top" align="start">
                    {department}
                  </TooltipContent>
                </Tooltip>
              ) : (
                <div className="truncate text-sm"><span className="text-muted-foreground">-</span></div>
              )
            ) : null}
            {jobLevel || workDate ? (
              <div className="flex min-w-0 items-center gap-1.5 overflow-hidden">
                {jobLevel ? (
                  <div className="w-[88px] shrink-0 overflow-hidden">
                    {jobLevel ? <Badge variant="secondary" className="max-w-full font-normal"><span className="truncate text-[13px]">{jobLevel}</span></Badge> : null}
                  </div>
                ) : null}
                {workDate ? (
                  <Badge variant="outline" className="min-w-0 gap-1 font-normal text-muted-foreground">
                    <CalendarDays size={13} className="shrink-0" aria-hidden="true" />
                    <span className="truncate text-[13px] tabular-nums">{workDate}</span>
                  </Badge>
                ) : null}
              </div>
            ) : null}
          </div>
        );
      },
    },
    dateColumn("lastSigninTime", i18next.t("user:Last signin time")),
    textColumn({dataIndex: "lastSigninIp", title: i18next.t("user:Last signin IP"), width: 150, searchable: true}),
    dateColumn(),
    organizationColumn(120),
    {
      dataIndex: "signupApplication",
      title: i18next.t("general:Application"),
      width: 140,
      sortable: true,
      searchable: true,
      render: (value, record) => value ? (
        <Link to={`/applications/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ) : null,
    },
    textColumn({dataIndex: "registerType", title: i18next.t("user:Register type"), width: 130, searchable: true}),
    textColumn({dataIndex: "registerSource", title: i18next.t("user:Register source"), width: 160, searchable: true}),
    boolColumn({dataIndex: "isAdmin", title: i18next.t("user:Is admin")}),
    {
      dataIndex: "pendingApproval",
      title: i18next.t("user:Pending approval"),
      width: 120,
      sortable: true,
      align: "center",
      render: (value) => (value ? <Badge variant="outline">{i18next.t("user:Pending approval")}</Badge> : null),
    },
    boolColumn({dataIndex: "isForbidden", title: i18next.t("user:Is forbidden"), invertColor: true}),
    boolColumn({dataIndex: "isDeleted", title: i18next.t("user:Is deleted"), invertColor: true}),
    {
      ...dateColumn("deletedTime", i18next.t("user:Deleted time")),
      render: (value, record) => record.isDeleted && value
        ? Setting.getFormattedDate(value)
        : <span className="text-muted-foreground">-</span>,
    },
  ];

  const displayColumns = columns.map((column): ColumnDef<any> => {
    if (["displayName", "department", "isAdmin", "isForbidden", "isDeleted", "deletedTime"].includes(column.dataIndex)) {
      return column;
    }
    return {
      ...column,
      render: (value, record, index) => {
        if (value == null || (typeof value === "string" && value.trim() === "")) {
          return <span className="text-muted-foreground">-</span>;
        }
        return column.render ? column.render(value, record, index) : value;
      },
    };
  });

  return (
    <CrudListPage
      title={i18next.t("general:Users")}
      description={groupName ? `${i18next.t("general:Groups")}: ${groupName}` : undefined}
      columns={displayColumns}
      formType="users"
      formItems={formItems}
      toolbar={Setting.isAdminUser(account) ? ({refresh}) => (
        <XlsxImport
          columns={Setting.getUserColumns()}
          templateName="import-user.xlsx"
          uploadApi="upload-users"
          successMessage="Users uploaded successfully, refreshing the page"
          onUploaded={refresh}
        />
      ) : undefined}
      deps={[organizationName, groupName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? UserBackend.getGlobalUsers(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
          : UserBackend.getUsers(
            organizationName,
            q.page,
            q.pageSize,
            q.searchedColumn,
            q.searchText,
            q.sortField,
            q.sortOrder,
            groupName,
          )
      }
      newRecord={account ? () => newUser(account, organization, organizationName, groupName) : undefined}
      deleteDisabled={isProtected}
      editUrl={(r) => `/users/${r.owner}/${r.name}`}
      remove={(r) => UserBackend.deleteUser(r)}
      rowActions={(record, _index, {refresh}) => [
        record.pendingApproval && Setting.isAdminUser(account)
          ? {
            key: "approve",
            label: i18next.t("user:Approve signup"),
            destructive: false,
            confirm: {title: i18next.t("user:Sure to approve"), description: `${record.name ?? ""}`},
            onSelect: () =>
              UserBackend.getUser(record.owner, record.name)
                .then((res: any) => {
                  if (res.status !== "ok" || !res.data) {
                    throw new Error(res.msg);
                  }
                  return UserBackend.updateUser(record.owner, record.name, {...res.data, isForbidden: false});
                })
                .then((res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("user:Successfully approved"));
                    refresh();
                  } else {
                    Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
                  }
                })
                .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${error.message ?? error}`)),
          }
          : null,
        Setting.isLocalAdminUser(account) && record.name !== account?.name
          ? {
            key: "impersonate",
            label: i18next.t("general:Impersonate"),
            onSelect: () => {
              UserBackend.impersonateUser(record.owner, record.name).then((res: any) => {
                if (res.status === "ok") {
                  navigate("/");
                  window.location.reload();
                } else {
                  Setting.showMessage("error", res.msg);
                }
              });
            },
          }
          : null,
        // only offered while the list is scoped to a group, as on the group tree page
        groupName
          ? {
            key: "remove",
            label: i18next.t("general:remove"),
            disabled: isProtected(record),
            // "remove from group", not "delete", so it asks its own question
            confirm: {title: i18next.t("general:Sure to remove"), description: `${record.name ?? ""}`},
            onSelect: () =>
              UserBackend.removeUserFromGroup({groupName, owner: record.owner, name: record.name})
                .then((res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("general:Successfully removed"));
                    refresh();
                  } else {
                    Setting.showMessage("error", `${i18next.t("general:Failed to remove")}: ${res.msg}`);
                  }
                })
                .catch((error) =>
                  Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`),
                ),
          }
          : null,
      ]}
      actionColumnWidth={groupName ? 400 : 330}
      // the tree page has one route per group, so pin the column choices to the page
      tableId={embedded ? "/trees" : undefined}
    />
  );
}
