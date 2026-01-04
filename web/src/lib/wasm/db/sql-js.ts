/**
 * WasmDatabase - sql.js wrapper with OPFS persistence.
 *
 * This provides a typed interface for interacting with the SQLite database
 * in the browser using sql.js (SQLite compiled to WebAssembly).
 */

import type { Database, SqlJsStatic } from 'sql.js'
import { getStorageBackend, type OPFSStorage, type IndexedDBStorage } from './opfs-storage'

// Lazy load sql.js to handle ESM import issues
async function loadSqlJs(): Promise<
  (config?: { locateFile?: (file: string) => string }) => Promise<SqlJsStatic>
> {
  const sqljs = await import('sql.js')

  // sql.js can export in different ways depending on bundler
  // Try different access patterns
  const initFn = sqljs.default?.default || sqljs.default || sqljs

  if (typeof initFn !== 'function') {
    throw new Error(
      `sql.js did not export a function. Got: ${typeof initFn}, keys: ${Object.keys(sqljs)}`
    )
  }

  return initFn as (config?: { locateFile?: (file: string) => string }) => Promise<SqlJsStatic>
}
// Note: Migrations are now handled exclusively by Go WASM (storage.RunMigrations)
// The schema.gen.ts file is kept for reference but not used here

export interface QueryResult<T> {
  columns: string[]
  values: T[]
}

export interface WasmDatabaseOptions {
  /** Path to sql.js WASM file. Defaults to '/wasm/sql-wasm.wasm' */
  wasmPath?: string
  /** Auto-persist interval in ms. Set to 0 to disable. Defaults to 30000 (30s) */
  autoPersistInterval?: number
}

export class WasmDatabase {
  private db: Database | null = null
  private SQL: SqlJsStatic | null = null
  private storage: OPFSStorage | IndexedDBStorage
  private wasmPath: string
  private autoPersistInterval: number
  private persistTimer: ReturnType<typeof setInterval> | null = null
  private dirty = false

  constructor(options: WasmDatabaseOptions = {}) {
    this.storage = getStorageBackend()
    this.wasmPath = options.wasmPath || '/wasm/sql-wasm.wasm'
    this.autoPersistInterval = options.autoPersistInterval ?? 30000
  }

  /**
   * Initialize the database.
   * Loads existing data from storage or creates a new database.
   *
   * Note: Migrations are handled by Go WASM (storage.RunMigrations),
   * not by TypeScript. This method only initializes sql.js and loads
   * any existing persisted data.
   */
  async initialize(): Promise<void> {
    // Load sql.js
    const initSqlJs = await loadSqlJs()
    this.SQL = await initSqlJs({
      locateFile: () => this.wasmPath,
    })

    // Try to load existing database or create empty one
    // Migrations are handled by Go WASM after this initialization
    const existingData = await this.storage.load()

    if (existingData) {
      this.db = new this.SQL.Database(existingData)
    } else {
      this.db = new this.SQL.Database()
    }

    // Start auto-persist timer
    if (this.autoPersistInterval > 0) {
      this.persistTimer = setInterval(() => {
        if (this.dirty) {
          this.persist().catch(console.error)
        }
      }, this.autoPersistInterval)
    }
  }

  /**
   * Persist database to storage.
   */
  async persist(): Promise<void> {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    const data = this.db.export()
    await this.storage.save(data)
    this.dirty = false
  }

  /**
   * Execute a SQL statement (no results returned).
   */
  exec(sql: string, params?: unknown[]): void {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    if (params && params.length > 0) {
      this.db.run(sql, params as (string | number | null | Uint8Array)[])
    } else {
      this.db.exec(sql)
    }

    this.dirty = true
  }

  /**
   * Execute a SQL query and return results.
   */
  query<T extends Record<string, unknown>>(sql: string, params?: unknown[]): T[] {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    const stmt = this.db.prepare(sql)

    if (params && params.length > 0) {
      stmt.bind(params as (string | number | null | Uint8Array)[])
    }

    const results: T[] = []
    while (stmt.step()) {
      const row = stmt.getAsObject() as T
      results.push(row)
    }
    stmt.free()

    return results
  }

  /**
   * Execute a SQL query and return a single result.
   */
  queryOne<T extends Record<string, unknown>>(sql: string, params?: unknown[]): T | null {
    const results = this.query<T>(sql, params)
    return results[0] || null
  }

  /**
   * Execute multiple SQL statements in a transaction.
   */
  transaction<T>(fn: () => T): T {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    this.db.exec('BEGIN TRANSACTION')
    try {
      const result = fn()
      this.db.exec('COMMIT')
      this.dirty = true
      return result
    } catch (error) {
      this.db.exec('ROLLBACK')
      throw error
    }
  }

  /**
   * Get the total number of rows in a table.
   */
  count(table: string, where?: string, params?: unknown[]): number {
    const whereClause = where ? ` WHERE ${where}` : ''
    const result = this.queryOne<{ count: number }>(
      `SELECT COUNT(*) as count FROM ${table}${whereClause}`,
      params
    )
    return result?.count || 0
  }

  /**
   * Insert a row and return the last inserted rowid.
   */
  insert(table: string, data: Record<string, unknown>): number {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    const columns = Object.keys(data)
    const placeholders = columns.map(() => '?').join(', ')
    const values = Object.values(data)

    this.exec(`INSERT INTO ${table} (${columns.join(', ')}) VALUES (${placeholders})`, values)

    // Get last inserted rowid
    const result = this.queryOne<{ id: number }>('SELECT last_insert_rowid() as id')
    return result?.id || 0
  }

  /**
   * Update rows in a table.
   */
  update(
    table: string,
    data: Record<string, unknown>,
    where: string,
    whereParams: unknown[]
  ): number {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    const sets = Object.keys(data)
      .map((col) => `${col} = ?`)
      .join(', ')
    const values = [...Object.values(data), ...whereParams]

    this.exec(`UPDATE ${table} SET ${sets} WHERE ${where}`, values)

    // Get number of affected rows
    const result = this.queryOne<{ changes: number }>('SELECT changes() as changes')
    return result?.changes || 0
  }

  /**
   * Delete rows from a table.
   */
  delete(table: string, where: string, params?: unknown[]): number {
    if (!this.db) {
      throw new Error('Database not initialized')
    }

    this.exec(`DELETE FROM ${table} WHERE ${where}`, params)

    const result = this.queryOne<{ changes: number }>('SELECT changes() as changes')
    return result?.changes || 0
  }

  /**
   * Close the database and clean up resources.
   */
  async close(): Promise<void> {
    if (this.persistTimer) {
      clearInterval(this.persistTimer)
      this.persistTimer = null
    }

    if (this.dirty) {
      await this.persist()
    }

    if (this.db) {
      this.db.close()
      this.db = null
    }
  }

  /**
   * Delete the database from storage and reset.
   */
  async reset(): Promise<void> {
    await this.close()
    await this.storage.delete()
    await this.initialize()
  }

  /**
   * Export the database as a Uint8Array.
   */
  export(): Uint8Array {
    if (!this.db) {
      throw new Error('Database not initialized')
    }
    return this.db.export()
  }

  /**
   * Import a database from a Uint8Array.
   */
  async import(data: Uint8Array): Promise<void> {
    if (!this.SQL) {
      throw new Error('sql.js not initialized')
    }

    // Close existing database
    if (this.db) {
      this.db.close()
    }

    // Create new database from data
    this.db = new this.SQL.Database(data)
    this.dirty = true

    // Persist immediately
    await this.persist()
  }

  /**
   * Check if the database is initialized.
   */
  isInitialized(): boolean {
    return this.db !== null
  }

  /**
   * Get database size in bytes.
   */
  getSize(): number {
    if (!this.db) {
      return 0
    }
    return this.db.export().byteLength
  }

  /**
   * Get the internal sql.js Database instance.
   * Used by Go WASM to share the same database.
   */
  getInternalDb(): Database | null {
    return this.db
  }

  /**
   * Get the sql.js SqlJsStatic instance.
   * Used by Go WASM to share the same sql.js.
   */
  getSqlJs(): SqlJsStatic | null {
    return this.SQL
  }
}

// Singleton instance
let dbInstance: WasmDatabase | null = null

/**
 * Get the singleton database instance.
 */
export function getDatabase(): WasmDatabase {
  if (!dbInstance) {
    dbInstance = new WasmDatabase()
  }
  return dbInstance
}

/**
 * Initialize the singleton database instance.
 */
export async function initializeDatabase(options?: WasmDatabaseOptions): Promise<WasmDatabase> {
  if (dbInstance?.isInitialized()) {
    return dbInstance
  }

  dbInstance = new WasmDatabase(options)
  await dbInstance.initialize()
  return dbInstance
}
