import * as React from "react"
import { CloudtrailRecord } from "../types"
import { HTMLTable } from "@blueprintjs/core"

type Props = {
  data: CloudtrailRecord[]
}

const CloudtrailRecords: React.FC<Props> = ({ data }) => (
  <HTMLTable interactive bordered striped>
    <thead>
      <tr>
        <th>Event Name</th>
        <th>Event Source</th>
      </tr>
    </thead>
    <tbody>
      {data.map((r, i) => (
        <tr style={{ marginBottom: 12 }} key={i}>
          <td>{r.eventName}</td>
          <td>{r.eventSource}</td>
        </tr>
      ))}
    </tbody>
  </HTMLTable>
)

export default CloudtrailRecords
