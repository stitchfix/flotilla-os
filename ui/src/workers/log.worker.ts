import { LogChunk } from "../types"

export default () => {
  onmessage = (evt: { data: { chunks: LogChunk[]; maxLen: number } }) => {
    const { chunks, maxLen } = evt.data
    let processed: string[] = []

    for (let i = 0; i < chunks.length; i++) {
      const { chunk } = chunks[i]
      const lines: string[] = chunk.split("\n")

      for (let j = 0; j < lines.length; j++) {
        const line = lines[j]

        if (line.length <= maxLen) {
          processed.push(line)
        } else {
          let k = 0

          while (k < line.length) {
            processed.push(line.substring(k, k + maxLen))
            k += maxLen
          }
        }
      }
    }

    postMessage(processed)
  }
}
