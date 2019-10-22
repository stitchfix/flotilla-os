import * as React from "react"
import { FormGroup, Classes } from "@blueprintjs/core"
import { FastField, FormikProps } from "formik"
import * as Yup from "yup"
import GroupNameSelect from "./GroupNameSelect"
import TagsSelect from "./TagsSelect"
import EnvFieldArray from "./EnvFieldArray"
import FieldError from "./FieldError"
import {
  groupNameFieldSpec,
  imageFieldSpec,
  commandFieldSpec,
  memoryFieldSpec,
  tagsFieldSpec,
  cpuFieldSpec,
} from "../constants"

export const validationSchema = {
  env: Yup.array().of(
    Yup.object().shape({
      name: Yup.string().required(),
      value: Yup.string().required(),
    })
  ),
  image: Yup.string()
    .min(1)
    .required("Required"),
  group_name: Yup.string()
    .min(1)
    .required("Required"),
  memory: Yup.number()
    .required("Required")
    .min(0),
  cpu: Yup.number()
    .required("Required")
    .min(0),
  command: Yup.string()
    .min(1)
    .required("Required"),
  tags: Yup.array().of(Yup.string()),
}

export type Props = Pick<
  FormikProps<any>,
  "values" | "setFieldValue" | "errors"
>

const BaseTaskForm: React.FunctionComponent<Props> = ({
  values,
  setFieldValue,
  errors,
}) => (
  <>
    <FormGroup
      label={groupNameFieldSpec.label}
      helperText={groupNameFieldSpec.description}
    >
      <FastField
        name={groupNameFieldSpec.name}
        component={GroupNameSelect}
        value={values.group_name}
        onChange={(value: string) => {
          setFieldValue(groupNameFieldSpec.name, value)
        }}
      />
      {errors.group_name && <FieldError>{errors.group_name}</FieldError>}
    </FormGroup>
    <FormGroup
      label={imageFieldSpec.label}
      helperText={imageFieldSpec.description}
    >
      <FastField name={imageFieldSpec.name} className={Classes.INPUT} />
      {errors.image && <FieldError>{errors.image}</FieldError>}
    </FormGroup>
    <FormGroup
      label={commandFieldSpec.label}
      helperText={commandFieldSpec.description}
    >
      <FastField
        className={`${Classes.INPUT} ${Classes.CODE}`}
        component="textarea"
        name={commandFieldSpec.name}
        rows={14}
      />
      {errors.command && <FieldError>{errors.command}</FieldError>}
    </FormGroup>
    <FormGroup label={cpuFieldSpec.label} helperText={cpuFieldSpec.description}>
      <FastField
        type="number"
        name={cpuFieldSpec.name}
        className={Classes.INPUT}
      />
      {errors.cpu && <FieldError>{errors.cpu}</FieldError>}
    </FormGroup>
    <FormGroup
      label={memoryFieldSpec.label}
      helperText={memoryFieldSpec.description}
    >
      <FastField
        type="number"
        name={memoryFieldSpec.name}
        className={Classes.INPUT}
      />
      {errors.memory && <FieldError>{errors.memory}</FieldError>}
    </FormGroup>
    <FormGroup
      label={tagsFieldSpec.label}
      helperText={tagsFieldSpec.description}
    >
      <FastField
        name={tagsFieldSpec.name}
        component={TagsSelect}
        value={values.tags}
        onChange={(value: string[]) => {
          setFieldValue(tagsFieldSpec.name, value)
        }}
      />
      {errors.tags && <FieldError>{errors.tags}</FieldError>}
    </FormGroup>
    <EnvFieldArray />
  </>
)

export default BaseTaskForm
