import * as React from "react"
import styled from "styled-components"
import JSONView from "react-json-view"
import Field from "./Field"
import { SPACING_PX, MONOSPACE_FONT_FAMILY } from "../../helpers/styles"
import Button from "./Button"
import ButtonGroup from "./ButtonGroup"

const KeyValuesContainer = styled.div`
  padding: ${SPACING_PX * 2}px;
`

const KeyValuesHeader = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  height: 100%;
  margin-bottom: ${SPACING_PX * 1.5}px;
`

interface IKeyValuesProps {
  actions?: React.ReactNode
  items: { [key: string]: React.ReactNode }
  label?: React.ReactNode
  raw: any
}

interface IKeyValuesState {
  displayRawData: boolean
}

class KeyValues extends React.PureComponent<IKeyValuesProps, IKeyValuesState> {
  static defaultProps: IKeyValuesProps = {
    items: {},
    raw: {},
  }

  state = {
    displayRawData: false,
  }

  onDisplayRawDataButtonClick = (evt: React.SyntheticEvent) => {
    this.setState(prevState => ({ displayRawData: !prevState.displayRawData }))
  }

  render() {
    const { actions, items, label, raw } = this.props
    const { displayRawData } = this.state

    let content

    if (!!displayRawData && !!raw) {
      content = (
        <JSONView
          src={raw}
          indentWidth={2}
          displayObjectSize={false}
          displayDataTypes={false}
          collapsed={1}
          theme={"ocean"}
          name={""}
          style={{
            fontSize: "0.9rem",
            fontWeight: 400,
            whiteSpace: "pre-wrap",
            wordBreak: "break-all",
            width: "100%",
            fontFamily: MONOSPACE_FONT_FAMILY,
          }}
        />
      )
    } else {
      content = Object.keys(items).map(key => (
        <Field label={key} key={key}>
          {items[key]}
        </Field>
      ))
    }

    return (
      <KeyValuesContainer>
        <KeyValuesHeader>
          <h3>{label}</h3>
          <ButtonGroup>
            <Button onClick={this.onDisplayRawDataButtonClick}>JSON</Button>
            {!!actions && actions}
          </ButtonGroup>
        </KeyValuesHeader>
        {content}
      </KeyValuesContainer>
    )
  }
}

export default KeyValues
