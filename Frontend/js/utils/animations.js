export function focusWithoutMotion(element) {
  element?.focus({ preventScroll: true });
}
