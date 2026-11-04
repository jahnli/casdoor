import * as React from "react";
import * as HoverCard from "@radix-ui/react-hover-card";
import i18next from "i18next";
import {Shield} from "lucide-react";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import * as Setting from "@/lib/setting";

export function UserProfileHoverCard({user, children}: {user: any; children: React.ReactNode}) {
  const primaryName = user.displayName || user.name;
  const avatarUrl = Setting.getEffectiveAvatarUrl(user);
  const properties = user.properties ?? {};
  const hasGender = user.gender === "Female" || user.gender === "Male";
  const fields = [
    {label: "user:Department name", chineseLabel: "部门", value: properties.department},
    {label: "user:Job level", chineseLabel: "岗位职级", value: properties.job_level},
    {label: "user:Work date", chineseLabel: "入职日期", value: properties.job_date},
    {label: "general:Email", chineseLabel: "邮箱", value: user.email},
    {label: "user:Job number", chineseLabel: "工号", value: properties.job_number},
    {label: "user:Job type", chineseLabel: "职务", value: properties.job_type},
    {label: "user:Job description", chineseLabel: "职位描述", value: properties.job_description},
    {label: "general:Phone", chineseLabel: "手机号", value: user.phone},
    {label: "user:Birthday", chineseLabel: "出生日期", value: user.birthday},
    {label: "user:Ethnicity", chineseLabel: "民族", value: properties.ethnicity},
    {label: "user:Location", chineseLabel: "探亲地", value: user.location},
  ].map((field) => ({...field, value: String(field.value ?? "").trim()}))
    .filter((field) => field.value !== "");

  return (
    <HoverCard.Root openDelay={100} closeDelay={150}>
      <HoverCard.Trigger asChild>{children}</HoverCard.Trigger>
      <HoverCard.Portal>
        <HoverCard.Content
          side="right"
          align="start"
          alignOffset={-96}
          sideOffset={4}
          collisionPadding={12}
          className="z-50 max-h-[var(--radix-hover-card-content-available-height)] w-80 max-w-[calc(100vw-24px)] overflow-y-auto rounded-xl border bg-popover p-0 text-popover-foreground shadow-lg outline-none data-[state=open]:animate-in data-[state=open]:fade-in-0"
        >
          <div
            className="h-28 rounded-t-lg bg-cover bg-center"
            style={properties.coverUrl
              ? {backgroundImage: `url(${properties.coverUrl})`}
              : {background: "linear-gradient(135deg, rgb(0, 90, 210) 0%, rgb(160, 210, 255) 100%)"}}
          />
          <div className="relative px-5 pb-3">
            <div className="-mt-8">
              <Avatar className="h-16 w-16 ring-[3px] ring-background">
                {avatarUrl ? <AvatarImage src={avatarUrl} alt={primaryName} /> : null}
                <AvatarFallback
                  className="text-base font-semibold text-white"
                  style={{backgroundColor: Setting.getAvatarColor(user.name)}}
                >
                  {(primaryName || "?").charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            </div>
            <div className="mt-2 flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-1.5">
                  <span className="truncate text-base font-semibold">{primaryName}</span>
                  {hasGender ? <span
                    aria-label={user.gender}
                    className={`shrink-0 text-sm ${user.gender === "Female" ? "text-pink-500" : "text-sky-500"}`}
                  >{user.gender === "Female" ? "♀" : "♂"}</span> : null}
                  {user.isAdmin ? <Shield size={14} className="shrink-0 text-primary" aria-label={i18next.t("user:Is admin")} /> : null}
                </div>
                {user.displayName && user.displayName !== user.name ? (
                  <div className="mt-0.5 truncate text-xs text-muted-foreground">@{user.name}</div>
                ) : null}
                {user.bio ? <div className="mt-1 truncate text-xs text-muted-foreground">{user.bio}</div> : null}
                {user.homepage ? <div className="mt-1 break-all text-xs text-muted-foreground">{user.homepage}</div> : null}
              </div>
              {user.affiliation ? <div className="mt-0.5 shrink-0 text-right text-xs text-muted-foreground">{user.affiliation}</div> : null}
            </div>
          </div>
          {fields.length > 0 ? (
            <>
              <div className="mx-5 h-px bg-border" />
              <div className="px-5 py-2.5">
                {fields.map((field) => (
                  <div key={field.label} className="flex items-baseline gap-3 py-1.5">
                    <span className="w-16 shrink-0 whitespace-nowrap text-[13px] leading-5 text-[#4b5672]">{Setting.getLanguage() === "zh" ? field.chineseLabel : i18next.t(field.label)}</span>
                    <span className="min-w-0 break-words text-sm leading-5 [overflow-wrap:anywhere]">{field.value}</span>
                  </div>
                ))}
              </div>
            </>
          ) : null}
        </HoverCard.Content>
      </HoverCard.Portal>
    </HoverCard.Root>
  );
}
