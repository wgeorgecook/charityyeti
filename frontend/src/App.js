import './style.css';
import Donate from './donate'
import Footer from './footer';
import Home from './home';
import {
  BrowserRouter as Router,
  Switch,
  Route
} from "react-router-dom";

function App() {
  return (
    <div className="App">
      <Router>
        <Switch>
          <Route path='/home'>
            <Home />
          </Route>
          <Route path='/donate'>
            <Donate />
          </Route>
          <Route path='/'>
            <Home />
          </Route>
        </Switch>
      </Router>
      <Footer />
    </div>
  );
}

export default App;
