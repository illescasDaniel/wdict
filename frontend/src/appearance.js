export const APPEARANCE_KEY = "wdict-appearance";
export const modes = ["system", "light", "dark"];

export function readAppearance(storage) {
  try {
    const value = storage.getItem(APPEARANCE_KEY);
    return modes.includes(value) ? value : "system";
  } catch {
    return "system";
  }
}

export function saveAppearance(storage, value) {
  try {
    storage.setItem(APPEARANCE_KEY, value);
  } catch {
    // The preference still applies for this session.
  }
}

export function resolveAppearance(value, systemDark) {
  return value === "system" ? (systemDark ? "dark" : "light") : value;
}
