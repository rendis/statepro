export interface NoteColorToken {
  bg: string;
  border: string;
  text: string;
  placeholder: string;
  scrollbarThumb: string;
}

export const NOTE_COLORS: NoteColorToken[] = [
  {
    bg: "bg-yellow-200",
    border: "border-yellow-400",
    text: "text-yellow-900",
    placeholder: "placeholder:text-yellow-700/50",
    scrollbarThumb: "rgba(161, 98, 7, 0.7)",
  },
  {
    bg: "bg-blue-200",
    border: "border-blue-400",
    text: "text-blue-900",
    placeholder: "placeholder:text-blue-700/50",
    scrollbarThumb: "rgba(29, 78, 216, 0.7)",
  },
  {
    bg: "bg-green-200",
    border: "border-green-400",
    text: "text-green-900",
    placeholder: "placeholder:text-green-700/50",
    scrollbarThumb: "rgba(21, 128, 61, 0.7)",
  },
  {
    bg: "bg-pink-200",
    border: "border-pink-400",
    text: "text-pink-900",
    placeholder: "placeholder:text-pink-700/50",
    scrollbarThumb: "rgba(190, 24, 93, 0.7)",
  },
  {
    bg: "bg-purple-200",
    border: "border-purple-400",
    text: "text-purple-900",
    placeholder: "placeholder:text-purple-700/50",
    scrollbarThumb: "rgba(109, 40, 217, 0.7)",
  },
];

export const MAX_NOTE_COLOR_INDEX = NOTE_COLORS.length - 1;

export const clampNoteColorIndex = (colorIndex: number | undefined): number => {
  if (!Number.isFinite(colorIndex)) {
    return 0;
  }

  return Math.min(MAX_NOTE_COLOR_INDEX, Math.max(0, Math.trunc(colorIndex || 0)));
};
