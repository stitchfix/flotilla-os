import * as React from "react"
import { isEmpty, isArray } from "lodash"
import { Env } from "../types"
import Attribute from "./Attribute"

const EnvList: React.FunctionComponent<{ env: Env[] }> = ({ env }) => (
  <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
    {isArray(env) &&
      !isEmpty(env) &&
      env.map(e => (
        <Attribute key={`${e.name}|${e.value}`} name={e.name} value={e.value} />
      ))}
  </div>
)

export default EnvList
