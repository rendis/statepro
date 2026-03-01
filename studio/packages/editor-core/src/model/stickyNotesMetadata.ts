import type { AnchoredNoteData, GlobalNoteData, JsonObject } from "../types";
import { clampNoteColorIndex } from "../components/canvas/notes";

export const STICKY_NOTE_METADATA_KEY = "_ui_note";
export const GLOBAL_NOTES_METADATA_KEY = "_ui_notes";

export interface StickyGlobalNoteSnapshot {
  x: number;
  y: number;
  data: GlobalNoteData;
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const toNoteText = (value: unknown): string => {
  if (typeof value !== "string") {
    return "";
  }

  return value;
};

const sanitizeAnchoredNote = (value: unknown): AnchoredNoteData | null => {
  if (!isRecord(value)) {
    return null;
  }

  return {
    text: toNoteText(value.text),
    colorIndex: clampNoteColorIndex(
      typeof value.colorIndex === "number" ? value.colorIndex : undefined,
    ),
  };
};

const sanitizeGlobalNote = (value: unknown): GlobalNoteData | null => {
  if (!isRecord(value)) {
    return null;
  }

  return {
    text: toNoteText(value.text),
    colorIndex: clampNoteColorIndex(
      typeof value.colorIndex === "number" ? value.colorIndex : undefined,
    ),
    isCollapsed: typeof value.isCollapsed === "boolean" ? value.isCollapsed : false,
  };
};

const cloneMetadata = (metadata: JsonObject | undefined): JsonObject => {
  if (!metadata) {
    return {};
  }

  return structuredClone(metadata);
};

export const stripStickyNoteKeysFromMetadata = (metadata: JsonObject): JsonObject => {
  const next = structuredClone(metadata);
  delete next[STICKY_NOTE_METADATA_KEY];
  delete next[GLOBAL_NOTES_METADATA_KEY];
  return next;
};

export const extractAnchoredNoteFromMetadata = (
  metadata: JsonObject | undefined,
): { metadata: JsonObject; note: AnchoredNoteData | null } => {
  const nextMetadata = cloneMetadata(metadata);
  const rawNote = nextMetadata[STICKY_NOTE_METADATA_KEY];
  delete nextMetadata[STICKY_NOTE_METADATA_KEY];
  delete nextMetadata[GLOBAL_NOTES_METADATA_KEY];

  return {
    metadata: nextMetadata,
    note: sanitizeAnchoredNote(rawNote),
  };
};

export const extractGlobalNotesFromMachineMetadata = (
  metadata: JsonObject | undefined,
): { metadata: JsonObject; notes: StickyGlobalNoteSnapshot[] } => {
  const nextMetadata = cloneMetadata(metadata);
  const rawNotes = nextMetadata[GLOBAL_NOTES_METADATA_KEY];

  delete nextMetadata[GLOBAL_NOTES_METADATA_KEY];
  delete nextMetadata[STICKY_NOTE_METADATA_KEY];

  if (!Array.isArray(rawNotes)) {
    return {
      metadata: nextMetadata,
      notes: [],
    };
  }

  const notes = rawNotes
    .map((raw): StickyGlobalNoteSnapshot | null => {
      if (!isRecord(raw)) {
        return null;
      }

      const x = typeof raw.x === "number" && Number.isFinite(raw.x) ? raw.x : null;
      const y = typeof raw.y === "number" && Number.isFinite(raw.y) ? raw.y : null;
      const data = sanitizeGlobalNote(raw);

      if (x === null || y === null || !data) {
        return null;
      }

      return {
        x,
        y,
        data,
      };
    })
    .filter((note): note is StickyGlobalNoteSnapshot => Boolean(note));

  return {
    metadata: nextMetadata,
    notes,
  };
};

export const injectAnchoredNoteIntoMetadata = (
  metadata: JsonObject,
  note: AnchoredNoteData | null | undefined,
): JsonObject => {
  const next = structuredClone(metadata);
  delete next[STICKY_NOTE_METADATA_KEY];

  if (note) {
    next[STICKY_NOTE_METADATA_KEY] = {
      text: toNoteText(note.text),
      colorIndex: clampNoteColorIndex(note.colorIndex),
    };
  }

  return next;
};

export const injectGlobalNotesIntoMetadata = (
  metadata: JsonObject,
  notes: StickyGlobalNoteSnapshot[],
): JsonObject => {
  const next = structuredClone(metadata);
  delete next[GLOBAL_NOTES_METADATA_KEY];

  if (notes.length > 0) {
    next[GLOBAL_NOTES_METADATA_KEY] = notes.map((note) => ({
      x: note.x,
      y: note.y,
      text: toNoteText(note.data.text),
      colorIndex: clampNoteColorIndex(note.data.colorIndex),
      isCollapsed: Boolean(note.data.isCollapsed),
    }));
  }

  return next;
};
