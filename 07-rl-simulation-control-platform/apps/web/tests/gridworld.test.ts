import { describe, expect, it } from 'vitest'

import { actionArrow, containsCoordinate, coordinateKey } from '../app/utils/gridworld'

describe('Gridworld presentation semantics', () => {
  it('uses stable coordinate keys', () => {
    expect(coordinateKey({ row: 5, column: 0 })).toBe('5,0')
  })

  it('detects registered obstacle coordinates', () => {
    expect(containsCoordinate([{ row: 2, column: 3 }], 2, 3)).toBe(true)
    expect(containsCoordinate([{ row: 2, column: 3 }], 3, 2)).toBe(false)
  })

  it('maps only known policy actions to visual arrows', () => {
    expect(actionArrow.right).toBe('→')
    expect(actionArrow.blocked).toBeUndefined()
  })
})
