import * as React from "react"

type Props = {
  children: (props: ChildProps) => React.ReactNode
}

type State = {
  isVisible: boolean
}

type ChildProps = {
  isVisible: boolean
  toggleVisibility: () => void
}

class Toggler extends React.Component<Props, State> {
  state = {
    isVisible: true,
  }

  toggleVisibility() {
    this.setState(prev => ({ isVisible: !prev.isVisible }))
  }

  getChildProps(): ChildProps {
    return {
      isVisible: this.state.isVisible,
      toggleVisibility: this.toggleVisibility.bind(this),
    }
  }

  render() {
    return this.props.children(this.getChildProps())
  }
}

export default Toggler
