import * as React from "react";
import i18next from "i18next";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as Conf from "@/Conf";
import * as SettingBackend from "@/backend/SettingBackend";

export default function SettingEditPage() {
  const [emailProviders, setEmailProviders] = React.useState<any[]>([]);

  React.useEffect(() => {
    ProviderBackend.getProviders("admin").then((res) => {
      if (res.status === "ok") {
        setEmailProviders((res.data ?? []).filter((provider) => provider.owner === "admin" && provider.category === "Email"));
      }
    });
  }, []);

  const updateSetting = async(record) => {
    const res = await SettingBackend.updateSetting(record);
    if (res.status === "ok") {
      Conf.setConfig({
        customHtmlAllowedTags: record.customHtmlAllowedTags ?? [],
        hasDefaultEmailProvider: !!record.defaultEmailProvider,
      });
    }
    return res;
  };

  const fields: EditField[] = [
    {type: "number", name: "dbMaxOpenConns", labelKey: "setting:Max open connections", min: 0},
    {
      type: "number",
      name: "dbMaxIdleConns",
      labelKey: "setting:Max idle connections",
      min: 0,
      validate: (value, ctx) =>
        ctx.record.dbMaxOpenConns > 0 && value > ctx.record.dbMaxOpenConns
          ? i18next.t("setting:Cannot be more than the max open connections")
          : undefined,
    },
    {
      type: "number",
      name: "dbConnMaxLifetime",
      labelKey: "setting:Connection max lifetime",
      min: 0,
      suffix: i18next.t("usage:seconds"),
    },
    {type: "number", name: "recordRetentionDays", labelKey: "setting:Default record retention days", min: 0},
    {
      type: "select",
      name: "defaultEmailProvider",
      labelKey: "setting:Default Email provider",
      options: () => [
        {value: "", label: i18next.t("general:None")},
        ...emailProviders.map((provider) => ({value: provider.name, label: provider.displayName ? `${provider.displayName} (${provider.name})` : provider.name})),
      ],
    },
    {type: "tags", name: "customHtmlAllowedTags", labelKey: "setting:Allowed custom HTML elements", placeholder: "iframe, video, source"},
    {type: "tags", name: "customHtmlAllowedUrls", labelKey: "setting:Allowed custom HTML URLs", placeholder: "https://www.youtube.com/embed/"},
    {type: "code", name: "headHtml", labelKey: "setting:Head HTML", language: "html"},
  ];

  return (
    <SimpleEditPage
      titleKey="general:Global Settings"
      backTo="/sysinfo"
      fields={fields}
      fetch={() => SettingBackend.getSetting()}
      add={updateSetting}
      update={updateSetting}
    />
  );
}
