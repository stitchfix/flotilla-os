import * as React from "react"
import { Callout, Intent } from "@blueprintjs/core"
import { get } from "lodash"
import { AxiosError } from "axios"
import Attribute from "./Attribute"

const ErrorCallout: React.FunctionComponent<{ error: AxiosError }> = ({
  error,
}) => {
  return (
    <Callout intent={Intent.DANGER}>
      <Attribute name="Code" value={error.code} />
      <Attribute name="Message" value={error.message} />
      <Attribute
        name="Response"
        value={get(error, ["response", "data", "error"], "")}
      />
    </Callout>
  )
}

export default ErrorCallout
