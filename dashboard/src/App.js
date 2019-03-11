import React from "react";
import RatingChart from "./RatingChart";

class App extends React.Component {
  constructor(props) {
    super(props);

    this.handleMessage = this.handleMessage.bind(this);
    this.addChartRow = this.addChartRow.bind(this);

    this.ws = new WebSocket("ws://localhost:8080/ws");

    this.ws.onopen = () => {
      this.setState({ connected: true });
    };
    this.ws.onerror = () => {
      console.log("ws error:", e);
    };
    this.ws.onmessage = e => {
      this.handleMessage(e);
    };
    this.ws.onclose = () => {
      this.setState({ connected: false });
    };
    this.ws.onopen = () => {
      this.setState({ connected: true });
    };

    this.state = {
      connected: false,
      chart_data: [["X", "Rating"]]
    };
  }
  addChartRow(row) {
    var arr = this.state.chart_data;
    arr.push([arr.length, row]);

    this.setState({ chart_data: arr });
  }
  handleMessage(e) {
    console.log(e.data);
    var obj = JSON.parse(e.data);

    switch (obj.command) {
      case "add_rating":
        obj.data.forEach(e => {
          this.addChartRow(e);
        });
        break;
      default:
        break;
    }
  }
  render() {
    return (
      <div>
        {this.state.connected && (
          <RatingChart ws={this.ws} chart_data={this.state.chart_data} />
        )}
      </div>
    );
  }
}

export default App;
