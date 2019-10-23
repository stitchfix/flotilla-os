import * as React from "react"
import { get } from "lodash"
import Creatable from "react-select/lib/Creatable"
import Request, { RequestStatus } from "./Request"
import { ListGroupsResponse, SelectOption, SelectProps } from "../types"
import api from "../api"
import * as helpers from "../helpers/selectHelpers"
import { Classes, Spinner } from "@blueprintjs/core"

/**
 * GroupNameSelect lets users choose a group name for their task definition. It
 * hits the `/groups` endpoint and renders the results into a React Select
 * component. If there are no existing groups, it will render an `<input>`
 * element as a fallback.
 */
export const GroupNameSelect: React.FunctionComponent<
  SelectProps & { options: SelectOption[] }
> = props => {
  return (
    <Creatable<SelectOption>
      value={helpers.stringToSelectOpt(props.value)}
      options={props.options}
      onChange={option => {
        props.onChange(helpers.preprocessSelectOption(option))
      }}
      isClearable
      id="groupNameSelect"
      styles={helpers.selectStyles}
      theme={helpers.selectTheme}
    />
  )
}

const ConnectedGroupNameSelect: React.FunctionComponent<
  SelectProps
> = props => (
  <Request<ListGroupsResponse, {}> requestFn={api.listGroups}>
    {({ data, requestStatus }) => {
      switch (requestStatus) {
        case RequestStatus.ERROR:
          return (
            <input
              className={Classes.INPUT}
              value={props.value}
              onChange={evt => {
                props.onChange(evt.target.value)
              }}
            />
          )
        case RequestStatus.READY:
          let options =
            get(data, "groups", []) === null ? [] : get(data, "groups", [])
          if (options === null) options = []
          return (
            <GroupNameSelect
              options={options.map(helpers.stringToSelectOpt)}
              value={props.value}
              onChange={props.onChange}
            />
          )
        case RequestStatus.NOT_READY:
        default:
          return <Spinner size={Spinner.SIZE_SMALL} />
      }
    }}
  </Request>
)

export default ConnectedGroupNameSelect
