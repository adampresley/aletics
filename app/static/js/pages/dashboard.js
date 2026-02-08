window.initDashboard = (viewOverTimeLabels, viewsOrderTimeData) => {
   // Replace the canvas element entirely to avoid stale dimensions from previous Chart.js instances
   const container = document.getElementById('pageViewsChartContainer');
   const oldCanvas = document.getElementById('pageViewsChart');
   if (oldCanvas) {
      const newCanvas = document.createElement('canvas');
      newCanvas.id = 'pageViewsChart';
      container.replaceChild(newCanvas, oldCanvas);
   }

   new Chart(document.getElementById('pageViewsChart'), {
      type: 'line',
      data: {
         labels: viewOverTimeLabels,
         datasets: [{
            label: 'Page Views',
            data: viewsOrderTimeData,
            fill: false,
            borderColor: 'rgb(75, 192, 192)',
            tension: 0.1
         }]
      },
      options: {
         responsive: true,
         maintainAspectRatio: false,
         scales: {
            y: {
               beginAtZero: true
            }
         }
      }
   });
};

