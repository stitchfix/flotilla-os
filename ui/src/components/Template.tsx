import * as React from "react"
import { Switch, Route, RouteComponentProps } from "react-router-dom"
import Request, { ChildProps, RequestStatus } from "./Request"
import api from "../api"
import { Template as TemplateShape } from "../types"
import TemplateDetails from "./TemplateDetails"
import TemplateRunForm from "./TemplateRunForm"

export type TemplateCtx = ChildProps<TemplateShape, { templateID: string }> & {
  basePath: string
  templateID: string
}

export const TemplateContext = React.createContext<TemplateCtx>({
  data: null,
  requestStatus: RequestStatus.NOT_READY,
  isLoading: false,
  error: null,
  request: () => {},
  basePath: "", // TODO: maybe this is not required.
  templateID: "",
  receivedAt: null,
})

export const Template: React.FunctionComponent<TemplateCtx> = props => {
  console.log(props.basePath)
  return (
    <TemplateContext.Provider value={props}>
      <Switch>
        <Route exact path={props.basePath} component={TemplateDetails} />
        <Route
          exact
          path={`${props.basePath}/execute`}
          component={TemplateRunForm}
        />
      </Switch>
    </TemplateContext.Provider>
  )
}

type ConnectedProps = RouteComponentProps<{ templateID: string }>
const Connected: React.FunctionComponent<ConnectedProps> = ({ match }) => (
  <Request<TemplateShape, { templateID: string }>
    requestFn={api.getTemplate}
    initialRequestArgs={{ templateID: match.params.templateID }}
  >
    {props => (
      <Template
        {...props}
        basePath={match.path}
        templateID={match.params.templateID}
      />
    )}
  </Request>
)

export default Connected
