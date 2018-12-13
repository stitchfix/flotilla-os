import * as React from "react"
import { get } from "lodash"
import TaskContext from "./TaskContext"
import View from "../styled/View"
import TaskHistoryTable from "./TaskHistoryTable"
import Navigation from "../Navigation/Navigation"
import DeleteTaskModal from "../Modal/DeleteTaskModal"
import ModalContext from "../Modal/ModalContext"
import {
  DetailViewContainer,
  DetailViewContent,
  DetailViewSidebar,
} from "../styled/DetailView"
import TaskDefinitionSidebar from "./TaskDefinitionSidebar"
import { requestStates, intents, IFlotillaUINavigationLink } from "../../.."

class UnwrappedTaskDefinition extends React.PureComponent<{
  renderModal: (modal: React.ReactNode) => void
}> {
  render() {
    return (
      <TaskContext.Consumer>
        {ctx => {
          const breadcrumbs = [
            { text: "Tasks", href: "/tasks" },
            {
              text: get(ctx, ["data", "alias"], ctx.definitionID),
              href: `/tasks/${ctx.definitionID}`,
            },
          ]
          let actions: IFlotillaUINavigationLink[] = []
          let sidebar = null

          switch (ctx.requestState) {
            case requestStates.READY:
              actions = [
                {
                  isLink: false,
                  text: "Delete",
                  buttonProps: {
                    intent: intents.ERROR,
                    onClick: () => {
                      this.props.renderModal(
                        <DeleteTaskModal definitionID={ctx.definitionID} />
                      )
                    },
                  },
                },
                {
                  isLink: true,
                  text: "Copy",
                  href: `/tasks/${ctx.definitionID}/copy`,
                },
                {
                  isLink: true,
                  text: "Edit",
                  href: `/tasks/${ctx.definitionID}/edit`,
                },
                {
                  isLink: true,
                  text: "Run",
                  href: `/tasks/${ctx.definitionID}/run`,
                  buttonProps: {
                    intent: intents.PRIMARY,
                  },
                },
              ]
              sidebar = <TaskDefinitionSidebar data={ctx.data} />
              break
            case requestStates.ERROR:
              sidebar = "blork"
              break
            case requestStates.NOT_READY:
            default:
              sidebar = "loading"
              break
          }

          return (
            <View>
              <Navigation breadcrumbs={breadcrumbs} actions={actions} />
              <DetailViewContainer>
                <DetailViewContent>
                  <TaskHistoryTable definitionID={ctx.definitionID} />
                </DetailViewContent>
                <DetailViewSidebar>{sidebar}</DetailViewSidebar>
              </DetailViewContainer>
            </View>
          )
        }}
      </TaskContext.Consumer>
    )
  }
}

const WrappedTaskDefinition: React.SFC<any> = () => (
  <ModalContext.Consumer>
    {ctx => <UnwrappedTaskDefinition renderModal={ctx.renderModal} />}
  </ModalContext.Consumer>
)

export default WrappedTaskDefinition
