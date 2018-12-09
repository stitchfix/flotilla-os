import { createGlobalStyle } from "styled-components"
import colors from "../../helpers/colors"
import { DEFAULT_FONT_COLOR } from "../../helpers/styles"

const GlobalStyle = createGlobalStyle`
  ::selection {
    background: ${colors.blue[4]};
    color: ${colors.gray[4]};
  }
  
  * {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }
  
  html {
    font-size: 14px;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  }

  body {
    background: ${colors.black[0]};
    font-size: 1rem;
    color: ${colors.gray[4]};
  }

  a {
    text-decoration: none;
    color: ${colors.blue[0]};
  }

  a.unstyled-link {
    color: ${DEFAULT_FONT_COLOR} !important;
    &:hover {
      color: ${colors.blue[0]} !important;
    }
  }

  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    font-weight: 300;
  }
  
  h1 {
    font-size: 2rem;
  }
  
  h2 {
    font-size: 1.5rem;
  }
  
  h3 {
    font-size: 1.25rem;
  }
  
  h4 {
    font-size: 1rem;
  }
  
  h5 {
    font-size: 0.875rem;
  }
  
  h6 {
    font-size: 0.75rem;
  }
`

export default GlobalStyle
