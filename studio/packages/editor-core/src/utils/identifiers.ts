const CHAR_DASH = 45;
const CHAR_UNDERSCORE = 95;
const CHAR_0 = 48;
const CHAR_9 = 57;
const CHAR_A = 65;
const CHAR_Z = 90;
const CHAR_a = 97;
const CHAR_z = 122;

const isAsciiLetter = (code: number): boolean =>
  (code >= CHAR_a && code <= CHAR_z) || (code >= CHAR_A && code <= CHAR_Z);

const isAsciiDigit = (code: number): boolean => code >= CHAR_0 && code <= CHAR_9;

const isWhitespaceChar = (ch: string): boolean => ch.trim() === "";

const trimTrailingChar = (value: string, code: number): string => {
  let end = value.length;
  while (end > 0 && value.charCodeAt(end - 1) === code) {
    end -= 1;
  }
  return value.slice(0, end);
};

export const formatIdentifier = (str: string): string => {
  const lower = str.toLowerCase();
  let out = "";
  let pendingDash = false;
  let started = false;

  for (let i = 0; i < lower.length; i += 1) {
    const ch = lower[i];
    const code = lower.charCodeAt(i);
    const letter = code >= CHAR_a && code <= CHAR_z;
    const digit = isAsciiDigit(code);
    const dashLike = code === CHAR_DASH || code === CHAR_UNDERSCORE || isWhitespaceChar(ch);

    if (!started) {
      if (letter) {
        started = true;
        out += ch;
      }
      continue;
    }

    if (letter || digit) {
      if (pendingDash) {
        out += "-";
        pendingDash = false;
      }
      out += ch;
      continue;
    }

    if (dashLike) {
      pendingDash = true;
    }
  }

  if (started && pendingDash) {
    out += "-";
  }

  return out;
};

export const cleanIdentifier = (str: string): string => {
  return trimTrailingChar(formatIdentifier(str), CHAR_DASH);
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
  const upper = str.toUpperCase();
  let out = "";
  let pendingUnderscore = false;

  for (let i = 0; i < upper.length; i += 1) {
    const ch = upper[i];
    const code = upper.charCodeAt(i);
    const letter = isAsciiLetter(code);
    const digit = isAsciiDigit(code);
    const underscoreLike = code === CHAR_UNDERSCORE || code === CHAR_DASH || isWhitespaceChar(ch);

    if (letter || digit) {
      if (pendingUnderscore && out.length > 0) {
        out += "_";
      }
      pendingUnderscore = false;
      out += ch;
      continue;
    }

    if (underscoreLike) {
      pendingUnderscore = true;
    }
  }

  if (out.length > 0 && pendingUnderscore) {
    out += "_";
  }

  return out;
};

export const cleanEventName = (str: string): string => {
  return trimTrailingChar(formatEventName(str), CHAR_UNDERSCORE);
};
