import type { Coordinate } from '~/types/api'

export const actionArrow: Record<string, string> = { up: '↑', right: '→', down: '↓', left: '←' }
export const coordinateKey = (point: Coordinate) => `${point.row},${point.column}`
export const containsCoordinate = (points: Coordinate[], row: number, column: number) =>
  points.some(point => point.row === row && point.column === column)
