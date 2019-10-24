import * as React from "react"
import { Button, FormGroup, Classes, Intent } from "@blueprintjs/core"
import { Env } from "../types"
import { IconNames } from "@blueprintjs/icons"
import { DebounceInput } from "react-debounce-input"
import { envFieldSpec } from "../helpers/taskFormHelpers"

type Props = {
  value: string[]
  onChange: (value: string[]) => void
}

type State = {
  newEnvName: string
  newEnvValue: string
}

class EnvQueryFilter extends React.Component<Props, State> {
  private delimiter: string = "|"

  constructor(props: Props) {
    super(props)
    this.handleNameChange = this.handleNameChange.bind(this)
    this.handleValueChange = this.handleValueChange.bind(this)
    this.handleRemove = this.handleRemove.bind(this)
    this.handleNewNameChange = this.handleNewNameChange.bind(this)
    this.handleNewValueChange = this.handleNewValueChange.bind(this)
    this.handleAddNewEnv = this.handleAddNewEnv.bind(this)
  }

  state = {
    newEnvName: "",
    newEnvValue: "",
  }

  serialize(env: Env): string {
    return `${env.name}${this.delimiter}${env.value}`
  }

  deserialize(str: string): Env {
    const split = str.split(this.delimiter)
    return {
      name: split[0],
      value: split[1],
    }
  }

  handleNameChange(i: number, evt: React.ChangeEvent<HTMLInputElement>) {
    const { value, onChange } = this.props
    const prevEnvValue = this.deserialize(value[i]).value
    const nextArr = value
    nextArr[i] = this.serialize({ name: evt.target.value, value: prevEnvValue })
    onChange(nextArr)
  }

  handleValueChange(i: number, evt: React.ChangeEvent<HTMLInputElement>) {
    const { value, onChange } = this.props
    const prevEnvName = this.deserialize(value[i]).name
    const nextArr = value
    nextArr[i] = this.serialize({ name: prevEnvName, value: evt.target.value })
    onChange(nextArr)
  }

  handleRemove(i: number) {
    const { value, onChange } = this.props
    let nextArr = value
    nextArr.splice(i, 1)
    onChange(nextArr)
  }

  handleNewNameChange(evt: React.ChangeEvent<HTMLInputElement>) {
    this.setState({ newEnvName: evt.target.value })
  }

  handleNewValueChange(evt: React.ChangeEvent<HTMLInputElement>) {
    this.setState({ newEnvValue: evt.target.value })
  }

  handleAddNewEnv() {
    const { value, onChange } = this.props
    const { newEnvName, newEnvValue } = this.state
    const prev = value
    const e = this.serialize({ name: newEnvName, value: newEnvValue })
    const next = prev.concat(e)
    this.setState({ newEnvName: "", newEnvValue: "" }, () => {
      onChange(next)
    })
  }

  render() {
    const { value } = this.props
    const { newEnvName, newEnvValue } = this.state

    return (
      <div>
        <div className="flotilla-env-field-array-header">
          <div className={Classes.LABEL}>{envFieldSpec.label}</div>
        </div>
        <div>
          {value.map((s: string, i: number) => {
            const e: Env = this.deserialize(s)
            return (
              <div key={i} className="flotilla-env-field-array-item">
                <FormGroup label={i === 0 ? "Name" : null}>
                  <DebounceInput
                    className={Classes.INPUT}
                    value={e.name}
                    onChange={this.handleNameChange.bind(this, i)}
                    debounceTimeout={500}
                  />
                </FormGroup>
                <FormGroup label={i === 0 ? "Value" : null}>
                  <DebounceInput
                    className={Classes.INPUT}
                    value={e.value}
                    onChange={this.handleValueChange.bind(this, i)}
                    debounceTimeout={500}
                  />
                </FormGroup>
                <Button
                  onClick={this.handleRemove.bind(this, i)}
                  type="button"
                  intent={Intent.DANGER}
                  style={i === 0 ? { transform: `translateY(8px)` } : {}}
                  icon={IconNames.CROSS}
                />
              </div>
            )
          })}
        </div>
        <div className="flotilla-env-field-array-item">
          <FormGroup label="Name">
            <input
              className={Classes.INPUT}
              value={newEnvName}
              onChange={this.handleNewNameChange}
            />
          </FormGroup>
          <FormGroup label="value">
            <input
              className={Classes.INPUT}
              value={newEnvValue}
              onChange={this.handleNewValueChange}
            />
          </FormGroup>
          <Button
            onClick={this.handleAddNewEnv}
            type="button"
            icon={IconNames.PLUS}
            style={{ transform: `translateY(8px)` }}
          />
        </div>
      </div>
    )
  }
}

export default EnvQueryFilter
