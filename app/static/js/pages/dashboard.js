window.renderJourneySankey = (containerId, edges) => {
   const container = document.getElementById(containerId);
   if (!container || !edges || edges.length === 0) return;

   container.innerHTML = '';

   const width = container.clientWidth || 900;
   const height = 420;
   const nodeW = 18;
   const nodePad = 10;
   const marginX = 170;
   const colors = ['#4e79a7','#f28e2b','#e15759','#76b7b2','#59a14f','#edc948','#b07aa1','#ff9da7','#9c755f','#bab0ac'];

   // Build nodes
   const nodes = {};
   const getNode = id => {
      if (!nodes[id]) nodes[id] = { id, inVal: 0, outVal: 0, col: -1 };
      return nodes[id];
   };
   edges.forEach(e => {
      getNode(e.source).outVal += e.value;
      getNode(e.target).inVal  += e.value;
   });

   // Assign columns: nodes with no incoming edge start at col 0
   Object.values(nodes).forEach(n => { if (n.inVal === 0) n.col = 0; });

   // Iteratively push targets one column past their source
   let changed = true;
   while (changed) {
      changed = false;
      edges.forEach(e => {
         const src = nodes[e.source], tgt = nodes[e.target];
         if (src.col >= 0 && tgt.col <= src.col) { tgt.col = src.col + 1; changed = true; }
      });
   }
   Object.values(nodes).forEach(n => { if (n.col < 0) n.col = 0; });

   const maxCol = Math.max(...Object.values(nodes).map(n => n.col));
   const numCols = maxCol + 1;
   const colXFn = c => marginX + c * ((width - 2 * marginX - nodeW) / Math.max(numCols - 1, 1));

   // Group and position nodes within each column
   const cols = {};
   for (let c = 0; c <= maxCol; c++) cols[c] = [];
   Object.values(nodes).forEach(n => cols[n.col].push(n));

   let colorIdx = 0;
   for (let c = 0; c <= maxCol; c++) {
      const col = cols[c];
      col.sort((a, b) => Math.max(b.inVal, b.outVal) - Math.max(a.inVal, a.outVal));
      const colTotal = col.reduce((s, n) => s + Math.max(n.inVal, n.outVal), 0) || 1;
      const usableH = height - nodePad * (col.length + 1);
      let y = nodePad;
      col.forEach(n => {
         n.color = colors[colorIdx++ % colors.length];
         n.x = colXFn(c);
         n.h = Math.max(24, (Math.max(n.inVal, n.outVal) / colTotal) * usableH);
         n.y = y;
         n.midY = y + n.h / 2;
         y += n.h + nodePad;
      });
   }

   const maxVal = Math.max(...edges.map(e => e.value));

   const svgNS = 'http://www.w3.org/2000/svg';
   const svg = document.createElementNS(svgNS, 'svg');
   svg.setAttribute('width', '100%');
   svg.setAttribute('height', height);
   svg.setAttribute('viewBox', `0 0 ${width} ${height}`);
   svg.style.fontFamily = 'inherit';

   // Draw edges first (behind nodes)
   edges.forEach(e => {
      const src = nodes[e.source], tgt = nodes[e.target];
      const sw = Math.max(2, Math.min(28, (e.value / maxVal) * 28));
      const x1 = src.x + nodeW, y1 = src.midY;
      const x2 = tgt.x,        y2 = tgt.midY;
      const cx = (x1 + x2) / 2;
      const path = document.createElementNS(svgNS, 'path');
      path.setAttribute('d', `M${x1},${y1} C${cx},${y1} ${cx},${y2} ${x2},${y2}`);
      path.setAttribute('fill', 'none');
      path.setAttribute('stroke', src.color);
      path.setAttribute('stroke-opacity', '0.35');
      path.setAttribute('stroke-width', sw);
      svg.appendChild(path);
   });

   // Draw nodes and labels
   Object.values(nodes).forEach(n => {
      const rect = document.createElementNS(svgNS, 'rect');
      rect.setAttribute('x', n.x);
      rect.setAttribute('y', n.y);
      rect.setAttribute('width', nodeW);
      rect.setAttribute('height', n.h);
      rect.setAttribute('fill', n.color);
      rect.setAttribute('rx', '3');
      svg.appendChild(rect);

      const isFirst = n.col === 0;
      const labelX = isFirst ? n.x - 6 : n.x + nodeW + 6;
      const label = n.id.length > 22 ? n.id.substring(0, 20) + '…' : n.id;
      const text = document.createElementNS(svgNS, 'text');
      text.setAttribute('x', labelX);
      text.setAttribute('y', n.midY + 4);
      text.setAttribute('text-anchor', isFirst ? 'end' : 'start');
      text.setAttribute('font-size', '11');
      text.setAttribute('fill', 'currentColor');
      text.textContent = label;
      svg.appendChild(text);
   });

   container.appendChild(svg);
};

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

