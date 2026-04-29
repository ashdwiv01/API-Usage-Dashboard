import Highcharts from "highcharts";
import HighchartsReact from "highcharts-react-official";

import { AggregatedMetric } from "../services/api";

type MetricsChartProps = {
  metrics: AggregatedMetric[];
  apiKey: string;
};

export function MetricsChart({ metrics, apiKey }: MetricsChartProps) {
  const requestsSeries = metrics.map((metric) => [metric.timestamp * 1000, metric.total]);
  const rejectedSeries = metrics.map((metric) => [metric.timestamp * 1000, metric.rejected]);

  const options: Highcharts.Options = {
    time: {
      timezone: "Asia/Kolkata",
    },
    chart: {
      backgroundColor: "transparent",
      height: 420,
      spacing: [24, 24, 24, 24],
    },
    title: {
      text: undefined,
    },
    credits: {
      enabled: false,
    },
    legend: {
      itemStyle: {
        color: "#d6d6cf",
        fontFamily: "IBM Plex Mono, monospace",
      },
    },
    xAxis: {
      type: "datetime",
      lineColor: "rgba(255,255,255,0.15)",
      tickColor: "rgba(255,255,255,0.15)",
      labels: {
        format: "{value:%H:%M:%S}",
        style: {
          color: "#a8b0aa",
          fontFamily: "IBM Plex Mono, monospace",
        },
      },
    },
    yAxis: {
      title: {
        text: "Requests / second",
        style: {
          color: "#a8b0aa",
          fontFamily: "IBM Plex Mono, monospace",
        },
      },
      gridLineColor: "rgba(255,255,255,0.08)",
      labels: {
        style: {
          color: "#a8b0aa",
          fontFamily: "IBM Plex Mono, monospace",
        },
      },
    },
    tooltip: {
      shared: true,
      xDateFormat: "%e %b %Y, %H:%M:%S IST",
      backgroundColor: "#111714",
      borderColor: "#35443d",
      style: {
        color: "#f6f3eb",
      },
    },
    plotOptions: {
      series: {
        animation: {
          duration: 300,
        },
        marker: {
          enabled: false,
        },
      },
    },
    series: [
      {
        type: "areaspline",
        name: `${apiKey || "API key"} requests/sec`,
        data: requestsSeries,
        color: "#5fd7b0",
        fillOpacity: 0.22,
      },
      {
        type: "spline",
        name: "Rejected (429)",
        data: rejectedSeries,
        color: "#ff7a59",
        dashStyle: "ShortDash",
        lineWidth: 3,
      },
    ],
  };

  return (
    <section className="panel chart-panel">
      <div className="panel-heading">
        <p className="eyebrow">Metrics Chart</p>
        <h2>Traffic pressure and rejection spikes</h2>
      </div>
      <HighchartsReact highcharts={Highcharts} options={options} />
    </section>
  );
}
