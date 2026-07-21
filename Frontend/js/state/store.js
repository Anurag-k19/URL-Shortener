const initialState = Object.freeze({ status: "idle", result: null, error: null });
let state = initialState;
const listeners = new Set();

export const store = Object.freeze({
  getState: () => state,
  setState(update) {
    state = Object.freeze({ ...state, ...update });
    listeners.forEach((listener) => listener(state));
  },
  subscribe(listener) {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },
  reset() {
    state = initialState;
    listeners.forEach((listener) => listener(state));
  },
});
