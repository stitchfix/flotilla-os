/**
 * Stores a chunk of logs and the chunk's lastSeen parameter.
 */
class LogChunk {
  constructor({ chunk, lastSeen }) {
    this.chunk = chunk
    this.lastSeen = lastSeen
  }

  getChunk = () => this.chunk
}

export default LogChunk
