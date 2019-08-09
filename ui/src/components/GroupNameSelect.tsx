import * as React from "react"
import { get, isArray } from "lodash"
import Creatable from "react-select/lib/Creatable"
import Request from "./Request"
import { ListGroupsResponse, SelectOption, SelectProps } from "../types"
import api from "../api"
import * as helpers from "../helpers/selectHelpers"

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
    />
  )
}

const ConnectedGroupNameSelect: React.FunctionComponent<
  SelectProps
> = props => (
  <Request<ListGroupsResponse, {}> requestFn={api.listGroups}>
    {res => {
      let options = get(res, ["data", "groups"], [])
      if (!isArray(options)) options = []
      return (
        <GroupNameSelect
          options={options.map(helpers.stringToSelectOpt)}
          value={props.value}
          onChange={props.onChange}
        />
      )
    }}
  </Request>
)

export default ConnectedGroupNameSelect
