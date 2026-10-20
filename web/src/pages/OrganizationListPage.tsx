import i18next from "i18next";
import dayjs from "dayjs";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import {newOrganization} from "@/pages/organization-defaults";

export default function OrganizationListPage({formItems}: {formItems?: any[]} = {}) {
  const organizationName = useOrganizationFilter();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (record) => `/organizations/${record.name}`, width: 140}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 160}),
    {
      dataIndex: "favicon",
      title: i18next.t("general:Favicon"),
      width: 70,
      align: "center",
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt="favicon" className="mx-auto h-8 w-8 object-contain" />
          </a>
        ) : null,
    },
    {
      dataIndex: "websiteUrl",
      title: i18next.t("organization:Website URL"),
      width: 200,
      sortable: true,
      searchable: true,
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
            {value}
          </a>
        ) : null,
    },
    textColumn({dataIndex: "passwordType", title: i18next.t("general:Password type"), width: 140, filters: valueFilters(["plain", "salt", "md5-salt"])}),
    textColumn({dataIndex: "passwordSalt", title: i18next.t("general:Password salt"), width: 130, searchable: true}),
    {
      dataIndex: "defaultAvatar",
      title: i18next.t("general:Default avatar"),
      width: 120,
      align: "center",
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt="avatar" className="mx-auto h-8 w-8 rounded-full object-cover" />
          </a>
        ) : null,
    },
    boolColumn({dataIndex: "enableSoftDeletion", title: i18next.t("organization:Soft deletion"), width: 120}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Organizations")}
      columns={columns}
      formType="organizations"
      formItems={formItems}
      rowKey={(row) => row.name}
      deps={[organizationName]}
      fetch={(query) =>
        OrganizationBackend.getOrganizations(
          "admin",
          organizationName,
          query.page,
          query.pageSize,
          query.searchedColumn,
          query.searchText,
          query.sortField,
          query.sortOrder,
        )
      }
      newRecord={() => newOrganization(dayjs().format())}
      // Groups and Users are the primary actions here, not Edit
      actionColumnWidth={330}
      rowActions={(record) => [
        {key: "groups", label: i18next.t("general:Groups"), href: `/trees/${record.name}`},
        {key: "users", label: i18next.t("general:Users"), href: `/organizations/${record.name}/users`},
      ]}
      // antd refuses to delete the built-in organization
      deleteDisabled={(record) => record.name === "built-in"}
      editUrl={(record) => `/organizations/${record.name}`}
      remove={(record) =>
        OrganizationBackend.deleteOrganization(record).then((res: any) => {
          if (res.status === "ok") {
            window.dispatchEvent(new Event("storageOrganizationsChanged"));
          }
          return res;
        })
      }
    />
  );
}
