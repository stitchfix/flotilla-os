import * as React from "react"
import { Card, Pre, Tag, Colors, Tooltip } from "@blueprintjs/core"
import { Run } from "../types"
import Attribute from "./Attribute"
import ISO8601AttributeValue from "./ISO8601AttributeValue"

const isLessThanPct = (x: number, y: number, pct: number): boolean => {
  if (x < pct * y) return true
  return false
}

const RunAttributes: React.FC<{ data: Run }> = ({ data }) => (
  <Card style={{ marginBottom: 12 }}>
    <div
      className="flotilla-attributes-container flotilla-attributes-container-horizontal"
      style={{ marginBottom: 12 }}
    >
      <Attribute name="Engine Type" value={<Tag>{data.engine}</Tag>} />
      <Attribute name="Cluster" value={data.cluster} />

      <Attribute
        name="Node Lifecycle"
        value={<Tag>{data.node_lifecycle || "-"}</Tag>}
      />
    </div>
    <div className="flotilla-form-section-divider" />
    <div
      className="flotilla-attributes-container flotilla-attributes-container-horizontal"
      style={{ marginBottom: 12 }}
    >
      <Attribute
        name="CPU (Units)"
        value={
          <div>
            <Tooltip content="Max CPU Used">
              <span
                style={{
                  color:
                    data.max_cpu_used &&
                    isLessThanPct(data.max_cpu_used, data.cpu, 0.5)
                      ? Colors.RED5
                      : "",
                }}
              >
                {data.max_cpu_used}
              </span>
            </Tooltip>{" "}
            / <Tooltip content="CPU Requested">{data.cpu}</Tooltip>
          </div>
        }
      />
      <Attribute
        name="Memory (MB)"
        value={
          <div>
            <Tooltip content="Max Memory Used">
              <span
                style={{
                  color:
                    data.max_memory_used &&
                    isLessThanPct(data.max_memory_used, data.memory, 0.5)
                      ? Colors.RED5
                      : "",
                }}
              >
                {data.max_memory_used}
              </span>
            </Tooltip>{" "}
            / <Tooltip content="Memory Requested">{data.memory}</Tooltip>
          </div>
        }
      />
    </div>
    {(data.ephemeral_storage || data.gpu) && (
      <div
        className="flotilla-attributes-container flotilla-attributes-container-horizontal"
        style={{ marginBottom: 12 }}
      >
        <Attribute
          name="Disk Size (GB)"
          value={data.ephemeral_storage || "-"}
        />
        <Attribute name="GPU Count" value={data.gpu || 0} />
      </div>
    )}
    <div className="flotilla-form-section-divider" />
    <div
      className="flotilla-attributes-container flotilla-attributes-container-horizontal"
      style={{ marginBottom: 12 }}
    >
      <Attribute
        name="Queued At"
        value={<ISO8601AttributeValue time={data.queued_at} />}
      />
      <Attribute
        name="Started At"
        value={<ISO8601AttributeValue time={data.started_at} />}
      />
      <Attribute
        name="Finished At"
        value={<ISO8601AttributeValue time={data.finished_at} />}
      />
    </div>
    <div className="flotilla-form-section-divider" />
    <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
      <Attribute
        name="Run ID"
        value={data.run_id}
        isCopyable
        rawValue={data.run_id}
      />
      <Attribute
        name="Definition ID"
        value={data.definition_id}
        isCopyable
        rawValue={data.definition_id}
      />
      <Attribute name="Image" value={data.image} />
      {data.pod_name && <Attribute name="EKS Pod Name" value={data.pod_name} />}
      <Attribute
        name="Command"
        value={
          data.command ? (
            <Pre className="flotilla-pre">
              {data.command.replace(/\n(\s)+/g, "\n")}
            </Pre>
          ) : (
            "Existing task definition command was used."
          )
        }
      />
    </div>
  </Card>
)

export default RunAttributes
