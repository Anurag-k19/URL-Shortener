const PREFIX = "snip:";

export const storage = Object.freeze({
  get(key) {
    const value = localStorage.getItem(`${PREFIX}${key}`);
    if (value === null) return null;
    try { return JSON.parse(value); } catch { return value; }
  },
  set(key, value) {
    localStorage.setItem(`${PREFIX}${key}`, JSON.stringify(value));
  },
  remove(key) {
    localStorage.removeItem(`${PREFIX}${key}`);
  },
});
