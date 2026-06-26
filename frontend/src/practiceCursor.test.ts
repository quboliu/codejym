import { describe, expect, it } from 'vitest'
import {
  advanceCursorForTypedChar,
  clampCursor,
  moveCursorBack,
  nextSourceLineCursor,
} from './practiceCursor'

describe('practice cursor model', () => {
  it('advances through comment text one typed character at a time', () => {
    const content = 'a // 中文注释\nb'
    let cursor = 2

    for (const char of '// 中文') {
      const result = advanceCursorForTypedChar(content, cursor, char)
      expect(result.matched).toBe(true)
      expect(result.cursor).toBe(cursor + char.length)
      cursor = result.cursor
    }

    expect(cursor).toBe(7)
  })

  it('does not jump across normal code after a matching character', () => {
    const content = 'const name = "CodeJYM"\nconsole.log(name)\n'
    const result = advanceCursorForTypedChar(content, 0, 'c')

    expect(result).toEqual({ cursor: 1, matched: true })
  })

  it('keeps cursor in place on mismatch', () => {
    const result = advanceCursorForTypedChar('abc', 1, 'z')

    expect(result).toEqual({ cursor: 1, matched: false })
  })

  it('moves back by one source character', () => {
    expect(moveCursorBack('abc', 3)).toBe(2)
    expect(moveCursorBack('你好吗', 2)).toBe(1)
  })

  it('skips only one source line, including comment lines', () => {
    const content = '// first comment\n// second comment\ncode\n'

    expect(nextSourceLineCursor(content, 0)).toBe('// first comment\n'.length)
  })

  it('skips to the next newline in the source, not to a visual wrap position', () => {
    const firstLine = 'const message = "this is a deliberately long source line";'
    const content = `${firstLine}\nnextLine()`

    expect(nextSourceLineCursor(content, 0)).toBe(firstLine.length + 1)
  })

  it('clamps restored server cursor to the source bounds', () => {
    expect(clampCursor('abc', -10)).toBe(0)
    expect(clampCursor('abc', 100)).toBe(3)
  })
})
