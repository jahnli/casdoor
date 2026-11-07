import type {CasdoorListResponse} from "@/components/crud/types";
import * as UserBackend from "@/backend/UserBackend";

export async function withAuditUsers<T extends Record<string, any>>(
  response: CasdoorListResponse<T> & {data3?: Record<string, any>},
  organizationField = "organization",
  userField = "user",
): Promise<CasdoorListResponse<T>> {
  if (response.status !== "ok" || !response.data?.length) {
    return response;
  }
  const users = new Map<string, Promise<any>>();
  const data = await Promise.all(response.data.map(async(row) => {
    const owner = row[organizationField];
    const name = row[userField];
    if (!owner || !name || name.startsWith("app/") || name.startsWith("app-dcr/")) {
      return row;
    }
    const id = `${owner}/${name}`;
    if (response.data3 && Object.prototype.hasOwnProperty.call(response.data3, id)) {
      return {...row, auditUser: response.data3[id] ?? {owner, name}};
    }
    const key = JSON.stringify([owner, name]);
    let pending = users.get(key);
    if (!pending) {
      pending = UserBackend.getUser(owner, name)
        .then((res) => res.status === "ok" ? res.data : null)
        .catch(() => null);
      users.set(key, pending);
    }
    const profile = await pending;
    return {...row, auditUser: profile ?? {owner, name}};
  }));
  return {...response, data};
}
