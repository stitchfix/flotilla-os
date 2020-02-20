import * as React from "react"
import { Link } from "react-router-dom"
import { get, omit } from "lodash"
import { Spinner, Callout } from "@blueprintjs/core"
import ListRequest, { ChildProps as ListRequestChildProps } from "./ListRequest"
import api from "../api"
import { ListTemplateParams, ListTemplateResponse, SortOrder } from "../types"
import pageToOffsetLimit from "../helpers/pageToOffsetLimit"
import Pagination from "./Pagination"
import ViewHeader from "./ViewHeader"
import { PAGE_SIZE } from "../constants"
import { RequestStatus } from "./Request"
import ErrorCallout from "./ErrorCallout"

export const initialQuery = {
  page: 1,
  sort_by: "template_name",
  order: SortOrder.ASC,
}

export type Props = ListRequestChildProps<
  ListTemplateResponse,
  { params: ListTemplateParams }
>

export const Templates: React.FunctionComponent<Props> = props => {
  const {
    data,
    updatePage,
    currentPage,
    isLoading,
    requestStatus,
    error,
  } = props

  let content: React.ReactNode

  switch (requestStatus) {
    case RequestStatus.ERROR:
      content = <ErrorCallout error={error} />
      break
    case RequestStatus.READY:
      if (data) {
        content = (
          <div className="flotilla-templates-container">
            {data.templates.map(t => (
              <Link
                className="flotilla-template-container"
                key={t.template_id}
                to={`/templates/${t.template_id}`}
              >
                <img
                  src={t.avatar_uri || ""}
                  width={36}
                  height={36}
                  alt="template-logo"
                />
                <div style={{ marginTop: 8 }}>
                  {t.template_name} v{t.version}
                </div>
              </Link>
            ))}
          </div>
        )
      } else {
        content = (
          <Callout>
            No templates found! Please contact your nearest Flotilla customer
            support agent for assistance.
          </Callout>
        )
      }
      break
    case RequestStatus.NOT_READY:
    default:
      content = <Spinner />
      break
  }

  return (
    <>
      <ViewHeader breadcrumbs={[{ text: "Templates", href: "/templates" }]} />
      <div className="flotilla-list-utils-container">
        <Pagination
          updatePage={updatePage}
          currentPage={currentPage}
          isLoading={isLoading}
          pageSize={PAGE_SIZE}
          numItems={data ? data.total : 0}
        />
      </div>
      {content}
    </>
  )
}

const ConnectedTasks: React.FunctionComponent = () => (
  <ListRequest<ListTemplateResponse, { params: ListTemplateParams }>
    requestFn={api.listTemplates}
    initialQuery={initialQuery}
    getRequestArgs={params => ({
      params: {
        ...omit(params, "page"),
        ...pageToOffsetLimit({
          page: get(params, "page", 1),
          limit: PAGE_SIZE,
        }),
      },
    })}
  >
    {props => <Templates {...props} />}
  </ListRequest>
)

export default ConnectedTasks
