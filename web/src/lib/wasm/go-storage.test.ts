import type { Database } from 'sql.js'
import { describe, expect, it, vi } from 'vitest'
import { createGoDatabaseBridge } from './go-storage'

describe('createGoDatabaseBridge', () => {
  it('marks the database dirty when Go runs a prepared write', () => {
    const markDirty = vi.fn()
    const statement = {
      run: vi.fn(function () {
        return this
      }),
    }
    const database = {
      prepare: vi.fn(() => statement),
    } as unknown as Database

    const bridge = createGoDatabaseBridge(database, markDirty)
    const bridgedStatement = bridge.prepare('INSERT INTO activities (id) VALUES (?)')

    expect(markDirty).not.toHaveBeenCalled()
    bridgedStatement.run([1])
    expect(markDirty).toHaveBeenCalledOnce()
  })
})
