import * as React from "react"
import { Card, Icon } from "@blueprintjs/core"
import urljoin from "url-join"
import { Run, ExecutionEngine } from "../types"
import Attribute from "./Attribute"

const createS3LogsUrl = (runID: string): string => {
  const prefix = process.env.REACT_APP_S3_LOGS_BUCKET_PREFIX || ""
  return urljoin(prefix, runID, "/")
}

const createEC2Url = (dns: string): string => {
  const prefix = process.env.REACT_APP_EC2_INSTANCE_URL_PREFIX || ""
  return `${prefix}${dns}`
}

const RunDebugAttributes: React.FC<{ data: Run }> = ({ data }) => (
  <Card style={{ marginTop: 12 }}>
    <div className="flotilla-card-header-container">
      <div className="flotilla-card-header">EKS Debug</div>
    </div>
    <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
      {data.pod_name && <Attribute name="EKS Pod Name" value={data.pod_name} />}
      {data.engine === ExecutionEngine.EKS && (
        <Attribute
          name="EKS S3 Logs"
          value={
            <a
              href={createS3LogsUrl(data.run_id)}
              target="_blank"
              rel="noopener noreferrer"
            >
              Link
              <Icon
                icon="share"
                style={{ marginLeft: 4, transform: "translateY(-2px)" }}
                iconSize={12}
              />
            </a>
          }
        />
      )}
      {data.instance.dns_name && (
        <Attribute
          name="EC2 Instance"
          value={
            <a
              href={createEC2Url(data.instance.dns_name)}
              target="_blank"
              rel="noopener noreferrer"
            >
              {data.instance.dns_name}
              <Icon
                icon="share"
                style={{ marginLeft: 4, transform: "translateY(-2px)" }}
                iconSize={12}
              />
            </a>
          }
        />
      )}
    </div>
  </Card>
)

export default RunDebugAttributes
