import * as React from "react"
import { Callout, Intent } from "@blueprintjs/core"
import { get } from "lodash"
import { AxiosError } from "axios"
import Attribute from "./Attribute"

const ErrorCallout: React.FunctionComponent<{ error: AxiosError | null }> = ({
  error,
}) => {
  return (
    <Callout intent={Intent.DANGER}>
      <div className="flotilla-attributes-container">
        <Attribute
          name="Code"
          value={error ? error.code : "No Error Code Provided"}
        />
        <Attribute
          name="Message"
          value={error ? error.message : "No Error Message Provided"}
        />
        <Attribute
          name="Response"
          value={get(error, ["response", "data", "error"], "")}
        />
      </div>
    </Callout>
  )
}

export default ErrorCallout
