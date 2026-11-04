/**
 * `organization.accountItems` drives which rows the user page shows and which of
 * them are editable. Ported from the `isAccountItemVisible` / `renderAccountItem`
 * gating in `web/src/UserEditPage.js`.
 *
 * UI visibility and editing follow the organization's policy; the backend also
 * filters protected fields and validates updates.
 */
export interface AccountItem {
  name: string;
  visible?: boolean;
  viewRule?: "Public" | "Self" | "Admin" | string;
  modifyRule?: "Self" | "Admin" | "Immutable" | string;
  tab?: string;
  regex?: string;
}

export interface AccountItemRules {
  isAdmin: boolean;
  isSelfOrAdmin: boolean;
  user: any;
}

export const ACCOUNT_PROPERTY_ITEMS = [
  {name: "Profile Cover", property: "coverUrl", labelKey: "user:Profile Cover", type: "url"},
  {name: "Open ID", property: "open_id", labelKey: "user:Open ID", type: "text"},
  {name: "Department name", property: "department", labelKey: "user:Department name", type: "text"},
  {name: "Department", property: "department_value", labelKey: "user:Department", type: "text"},
  {name: "Job number", property: "job_number", labelKey: "user:Job number", type: "text"},
  {name: "Work date", property: "job_date", labelKey: "user:Work date", type: "date"},
  {name: "Job type", property: "job_type", labelKey: "user:Job type", type: "text"},
  {name: "Job level", property: "job_level", labelKey: "user:Job level", type: "text"},
];

export const ACCOUNT_ITEM_ORDER = [
  "Avatar", "Organization", "Display name", "Name", "Last name", "First name", "Email", "Phone", "Country code",
  "Groups", "Affiliation", "Department name", "Department", "Job number", "Job type", "Job level", "Work date",
  "Gender", "Birthday", "Education", "Language", "Title", "Homepage", "Profile Cover", "Bio",
  "Signup application", "Register type", "Register source", "ID", "Open ID",
  "User type", "Tag", "Is admin", "Is forbidden", "Roles", "Permissions", "Password", "Need update password", "Last change password time",
  "UID number", "IP whitelist", "Last signin time", "Last signin IP", "Is deleted", "Deleted time", "Properties", "Country/Region", "Location", "Address", "Addresses", "Consents",
  "Multi-factor authentication", "WebAuthn credentials", "Face ID", "3rd-party logins", "MFA accounts", "MFA items", "Managed accounts",
  "Real name", "ID card type", "ID card", "ID card info", "ID verification",
];

/** The rows whose editing is frozen once the user's identity has been verified. */
const ID_VERIFICATION_ITEMS = new Set(["ID card info", "ID card", "ID card type", "Real name"]);

export function getAccountItemMap(organization: any): Record<string, AccountItem> {
  const items: AccountItem[] = organization?.accountItems ?? [];
  const map: Record<string, AccountItem> = {};
  items.forEach((item) => {
    if (item?.name) {
      map[item.name] = item;
    }
  });
  return map;
}

export function isAccountItemVisible(item: AccountItem | undefined, rules: AccountItemRules): boolean {
  if (!item || !item.visible) {
    return false;
  }
  if (item.viewRule === "Self") {
    return rules.isSelfOrAdmin;
  }
  if (item.viewRule === "Admin") {
    return rules.isAdmin;
  }
  return true;
}

export function isAccountItemDisabled(item: AccountItem | undefined, rules: AccountItemRules): boolean {
  if (!item) {
    return false;
  }

  let disabled = false;
  if (item.modifyRule === "Self") {
    disabled = !rules.isSelfOrAdmin;
  } else if (item.modifyRule === "Admin") {
    disabled = !rules.isAdmin;
  } else if (item.modifyRule === "Immutable") {
    disabled = true;
  }

  // the built-in admin cannot be moved between organizations or renamed
  if (item.name === "Organization" || item.name === "Name") {
    if (rules.user?.owner === "built-in" && rules.user?.name === "admin") {
      disabled = true;
    }
  }

  if (ID_VERIFICATION_ITEMS.has(item.name) && rules.user?.isVerified) {
    disabled = true;
  }

  return disabled;
}

/** `true` when the row's `regex` rejects the current value, like antd's Form rule. */
export function checkAccountItemRegex(item: AccountItem | undefined, value: any): boolean {
  if (!item?.regex || value === undefined || value === null || value === "") {
    return false;
  }
  try {
    return !new RegExp(item.regex).test(String(value));
  } catch (error) {
    // an organization can save an invalid regex; antd would throw here, we just pass
    return false;
  }
}
