export default () => {
  onmessage = (evt: { data: { logs: string; maxLen: number } }) => {
    const { logs, maxLen } = evt.data
    let processed: string[] = []

    // Split `logs` string by newline char.
    const lines: string[] = logs.split("\n")

    // Iterate over each line. If line.length <= maxLen, push to `processed`
    // array. If the length of the line is greater than maxLen, iterate over
    // the line `maxLen` chars at a time and push each sub-line to the
    // `processed` array.
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

    postMessage(processed)
  }
}
