import { HitStore, HitEvent, HitQuery, EndpointSummary, TotalSummary } from './base'

export class AsyncWriteBuffer extends HitStore {
  private queue: HitEvent[] = []
  private timer: ReturnType<typeof setInterval> | null = null

  constructor(
    private readonly backend: HitStore,
    private readonly flushInterval = 2000,
    private readonly maxBuffer = 10_000,
    private readonly batchSize = 100,
  ) {
    super()
  }

  start(): void {
    this.timer = setInterval(() => { void this._flush() }, this.flushInterval)
    // Don't keep the Node.js process alive just for the flush timer
    this.timer.unref()
  }

  async record(hit: HitEvent): Promise<void> {
    if (this.queue.length >= this.maxBuffer) {
      this.queue.shift() // drop oldest on overflow
    }
    this.queue.push(hit)
  }

  private async _flush(): Promise<void> {
    const batch = this.queue.splice(0, this.batchSize)
    for (const hit of batch) {
      await this.backend.record(hit)
    }
  }

  /** Flush all pending hits synchronously — useful in tests and graceful shutdown. */
  async flushAll(): Promise<void> {
    while (this.queue.length > 0) {
      await this._flush()
    }
  }

  async recentHits(query?: HitQuery): Promise<HitEvent[]> {
    return this.backend.recentHits(query)
  }

  async endpointSummaries(since_days?: number): Promise<EndpointSummary[]> {
    return this.backend.endpointSummaries(since_days)
  }

  async totalSummary(since_days?: number): Promise<TotalSummary> {
    return this.backend.totalSummary(since_days)
  }

  async close(): Promise<void> {
    if (this.timer) clearInterval(this.timer)
    await this.flushAll()
    await this.backend.close()
  }
}
