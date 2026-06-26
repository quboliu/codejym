export interface TypedCharResult {
  cursor: number
  matched: boolean
}

export function clampCursor(content: string, cursor: number): number {
  if (!Number.isFinite(cursor)) return 0
  return Math.max(0, Math.min(content.length, Math.trunc(cursor)))
}

export function advanceCursorForTypedChar(content: string, cursor: number, typedChar: string): TypedCharResult {
  const safeCursor = clampCursor(content, cursor)
  if (!typedChar || safeCursor >= content.length) {
    return { cursor: safeCursor, matched: false }
  }

  const expected = content.slice(safeCursor, safeCursor + typedChar.length)
  if (expected !== typedChar) {
    return { cursor: safeCursor, matched: false }
  }

  return {
    cursor: clampCursor(content, safeCursor + typedChar.length),
    matched: true,
  }
}

export function moveCursorBack(content: string, cursor: number): number {
  const safeCursor = clampCursor(content, cursor)
  if (safeCursor === 0) return 0
  const previousChar = Array.from(content.slice(0, safeCursor)).at(-1)
  return clampCursor(content, safeCursor - (previousChar?.length ?? 1))
}

export function nextSourceLineCursor(content: string, cursor: number): number {
  const safeCursor = clampCursor(content, cursor)
  if (safeCursor >= content.length) return content.length
  const newlineIndex = content.indexOf('\n', safeCursor)
  return newlineIndex === -1 ? content.length : newlineIndex + 1
}
