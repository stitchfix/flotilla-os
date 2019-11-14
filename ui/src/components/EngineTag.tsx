import * as React from "react"
import { Tag } from "@blueprintjs/core"
import { ExecutionEngine } from "../types"

const EngineTag: React.FC<{ engine: ExecutionEngine }> = ({ engine }) => (
  <Tag>{engine}</Tag>
)

export default EngineTag
