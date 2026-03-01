export const formatIdentifier = (str: string): string => {
  return str
    .toLowerCase()
    .replace(/[\s_]+/g, "-")
    .replace(/[^a-z0-9-]/g, "")
    .replace(/-+/g, "-")
    .replace(/^[^a-z]+/, "");
};

export const cleanIdentifier = (str: string): string => {
  return formatIdentifier(str).replace(/-+$/, "");
};

export const ensureUniqueIdentifier = (
  base: string,
  usedValues: Set<string>,
  fallback = "item",
): string => {
  const cleanBase = cleanIdentifier(base) || cleanIdentifier(fallback) || "item";
  if (!usedValues.has(cleanBase)) {
    return cleanBase;
  }

  let suffix = 2;
  while (usedValues.has(`${cleanBase}-${suffix}`)) {
    suffix += 1;
  }

  return `${cleanBase}-${suffix}`;
};

export const formatEventName = (str: string): string => {
  return str
    .toUpperCase()
    .replace(/[\s\-]+/g, "_")
    .replace(/[^A-Z0-9_]/g, "")
    .replace(/_+/g, "_")
    .replace(/^_/, "");
};

export const cleanEventName = (str: string): string => {
  return formatEventName(str).replace(/_$/, "");
};
