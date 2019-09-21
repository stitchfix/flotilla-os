import * as React from "react"
import { Link } from "react-router-dom"
import { Breadcrumbs, IBreadcrumbProps, Classes } from "@blueprintjs/core"

type Props = {
  breadcrumbs: IBreadcrumbProps[]
  buttons?: React.ReactNode
}

const ViewHeader: React.FunctionComponent<Props> = ({
  breadcrumbs,
  buttons,
}) => (
  <div className="flotilla-view-header-container">
    <Breadcrumbs
      items={breadcrumbs}
      breadcrumbRenderer={(props: IBreadcrumbProps) => (
        <Link to={props.href ? props.href : "/"}>{props.text}</Link>
      )}
      className={Classes.TEXT_LARGE}
    />
    {buttons}
  </div>
)

export default ViewHeader
