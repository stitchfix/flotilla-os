import * as React from "react"
import { useSelector, useDispatch } from "react-redux"
import { Formik, Form, FastField, Field } from "formik"
import {
  Classes,
  Button,
  Dialog,
  Switch,
  FormGroup,
  Intent,
} from "@blueprintjs/core"
import { RootState } from "../state/store"
import {
  Settings,
  update,
  toggleDialogVisibilityChange,
} from "../state/settings"

const SettingsButton: React.FC = () => {
  const dispatch = useDispatch()
  const { settings, isSettingsDialogOpen, isLoading } = useSelector(
    (s: RootState) => s.settings
  )

  return (
    <>
      <Button
        rightIcon="cog"
        onClick={() => {
          dispatch(toggleDialogVisibilityChange(true))
        }}
      >
        Settings
      </Button>
      <Dialog
        isOpen={isSettingsDialogOpen}
        onClose={() => {
          dispatch(toggleDialogVisibilityChange(false))
        }}
        className="bp3-dark"
        title="Settings"
      >
        <Formik<Settings>
          initialValues={settings}
          onSubmit={values => {
            dispatch(update(values))
          }}
        >
          {({ values, setFieldValue }) => {
            return (
              <Form>
                <div className={Classes.DIALOG_BODY}>
                  <FormGroup helperText="Enabling this will ensure that the UI doesn't crash for runs with massive log output">
                    <FastField
                      name="USE_OPTIMIZED_LOG_RENDERER"
                      component={Switch}
                      checked={values.USE_OPTIMIZED_LOG_RENDERER}
                      onChange={() => {
                        setFieldValue(
                          "USE_OPTIMIZED_LOG_RENDERER",
                          !values.USE_OPTIMIZED_LOG_RENDERER
                        )
                      }}
                      label="Use optimized log renderer."
                    />
                  </FormGroup>
                  <FormGroup helperText="Enabling this will allow you to search through the optimized logs by pressing ⌘-F">
                    <Field
                      name="SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW"
                      component={Switch}
                      checked={values.SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW}
                      onChange={() => {
                        setFieldValue(
                          "SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW",
                          !values.SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW
                        )
                      }}
                      label="Override ⌘-F in run view."
                      disabled={values.USE_OPTIMIZED_LOG_RENDERER === false}
                    />
                  </FormGroup>
                </div>
                <div className={Classes.DIALOG_FOOTER}>
                  <div className={Classes.DIALOG_FOOTER_ACTIONS}>
                    <Button
                      onClick={() => {
                        dispatch(toggleDialogVisibilityChange(false))
                      }}
                    >
                      Close
                    </Button>
                    <Button
                      intent={Intent.PRIMARY}
                      type="submit"
                      loading={isLoading}
                    >
                      Save Changes
                    </Button>
                  </div>
                </div>
              </Form>
            )
          }}
        </Formik>
      </Dialog>
    </>
  )
}

export default SettingsButton
