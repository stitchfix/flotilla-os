import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { get } from "lodash"
import { TaskFormNavigation, IProps } from "../TaskFormNavigation"
import { flotillaUIIntents, IFlotillaUINavigationLink } from "../../../types"

const goBack = jest.fn()
const defaultProps: IProps = {
  isSubmitDisabled: true,
  inFlight: false,
  breadcrumbs: [],
  goBack,
}

describe("TaskFormNavigation", () => {
  it("renders a Navigation component", () => {
    const wrapper = mount(
      <MemoryRouter>
        <TaskFormNavigation {...defaultProps} />
      </MemoryRouter>
    )

    const navigationWrapper = wrapper.find("Navigation")

    expect(navigationWrapper.length).toBe(1)
    expect(navigationWrapper.prop("breadcrumbs")).toBe(defaultProps.breadcrumbs)

    // const actionsProp = navigationWrapper.prop(
    //   "actions"
    // ) as IFlotillaUINavigationLink[]

    // expect(actionsProp[0].isLink).toBe(false)
    // expect(actionsProp[0].text).toBe("Cancel")
    // expect(actionsProp[0].buttonProps).toHaveProperty("onClick")
    // expect(get(actionsProp[0].buttonProps, "onClick")).toEqual(
    //   defaultProps.goBack
    // )
    expect(navigationWrapper.prop("actions")).toEqual([
      {
        isLink: false,
        text: "Cancel",
        buttonProps: {
          onClick: defaultProps.goBack,
        },
      },
      {
        isLink: false,
        text: "Submit",
        buttonProps: {
          type: "submit",
          intent: flotillaUIIntents.PRIMARY,
          isDisabled: defaultProps.isSubmitDisabled,
          isLoading: !!defaultProps.inFlight,
        },
      },
    ])
  })
})
