/**
 * The screen sizes the layout is checked at (#359): one list for the
 * deterministic checks (sizes.spec.ts) and the ui-review screens
 * (ui-shots.spec.ts). Adding a row runs every check at that size too. The
 * width bands are those of #353 and the layout table in
 * docs/dev/conventions.md.
 */
export const SIZES = [
  { name: 'desktop', width: 1280, height: 800 },
  { name: 'phone', width: 390, height: 844 },
  { name: 'half', width: 960, height: 1080 },
  { name: 'fhd', width: 1920, height: 1080 },
  { name: 'qhd', width: 2560, height: 1440 },
] as const;
