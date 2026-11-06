import * as React from "react";
import {Link} from "react-router-dom";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {UserProfileHoverCard} from "@/components/user/UserProfileHoverCard";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

const pendingUsers = new Map<string, Promise<any>>();

function loadUser(organization: string, name: string) {
  const key = JSON.stringify([organization, name]);
  let pending = pendingUsers.get(key);
  if (!pending) {
    pending = UserBackend.getUser(organization, name)
      .then((res) => res.status === "ok" ? res.data : null)
      .catch(() => null)
      .finally(() => pendingUsers.delete(key));
    pendingUsers.set(key, pending);
  }
  return pending;
}

export function UserCell({user, organization, name, linkToUser = false}: {
  user?: any;
  organization?: string;
  name?: string;
  linkToUser?: boolean;
}) {
  const [loaded, setLoaded] = React.useState<{organization: string; name: string; user: any} | null>(null);

  React.useEffect(() => {
    if (user || !organization || !name) {
      return;
    }
    let active = true;
    loadUser(organization, name).then((result) => {
      if (active) {
        setLoaded({organization, name, user: result});
      }
    });
    return () => { active = false; };
  }, [user, organization, name]);

  const profile = user ?? (loaded?.organization === organization && loaded?.name === name ? loaded.user : null);
  const record = profile ?? {owner: organization, name};
  if (!record.name) {
    return null;
  }

  const url = Setting.getEffectiveAvatarUrl(record);
  const avatar = (
    <Avatar className="h-7 w-7 shrink-0 text-xs">
      {url ? <AvatarImage src={url} alt={record.name} /> : null}
      <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(record.name), color: "#fff"}}>
        {record.name.charAt(0).toUpperCase()}
      </AvatarFallback>
    </Avatar>
  );
  const names = (
    <>
      <span>{record.displayName}</span>
      <span className="text-[13px] text-[#4b5672]">{record.name}</span>
    </>
  );

  return (
    <div className="flex w-fit max-w-full items-center gap-3">
      <UserProfileHoverCard user={record}>
        {url ? (
          <a href={url} target="_blank" rel="noreferrer" className="shrink-0">{avatar}</a>
        ) : (
          <span className="shrink-0 cursor-pointer" tabIndex={0}>{avatar}</span>
        )}
      </UserProfileHoverCard>
      {linkToUser ? (
        <Link to={`/users/${record.owner}/${encodeURIComponent(record.name)}`} className="flex min-w-0 flex-col underline-offset-4 hover:underline">
          {names}
        </Link>
      ) : (
        <div className="flex min-w-0 flex-col">{names}</div>
      )}
    </div>
  );
}
