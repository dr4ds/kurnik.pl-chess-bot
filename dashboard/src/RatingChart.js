import React from "react";
import Chart from "react-google-charts";
import { buildPayload } from "./utils";

class RatingChart extends React.Component {
  constructor(props) {
    super(props);

    this.props.ws.send(buildPayload("init_rating", null));
  }

  render() {
    return (
      <div>
        {this.props.chart_data.length > 1 && (
          <Chart
            chartType="LineChart"
            width="60%"
            height="200px"
            data={this.props.chart_data}
          />
        )}
      </div>
    );
  }
}

export default RatingChart;
