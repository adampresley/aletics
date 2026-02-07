(() => {
   if (window.Aletics) {
      return;
   }

   function assemble() {
      return {
         path: window.location.pathname,
         queryString: window.location.search,
         browser: getBrowserName(),
      };
   }

   function getBrowserName() {
      const ua = navigator.userAgent;

      if (ua.includes("Firefox")) return "Firefox";
      if (ua.includes("Edg")) return "Edge";
      if (ua.includes("Chrome")) return "Chrome";
      if (ua.includes("Safari")) return "Safari";

      return "Unknown";
   }

   function newTracker(baseUrl) {
      const t = {
         track: () => {
            const payload = assemble();
            const data = JSON.stringify(payload);

            if (navigator.sendBeacon) {
               navigator.sendBeacon(`${baseUrl}/v1/track`, data);
            } else {
               fetch(`${baseUrl}/v1/track`, {
                  method: "POST",
                  "Content-Type": "text/plain;charset=UTF-8",
                  body: data,
                  keepalive: true,
               }).catch(err => console.error(`Aletics tracking error:`, err));
            }
         },
      };

      return t;
   }

   const Aletics = {
      init: (baseUrl) => {
         return newTracker(baseUrl);
      },
   };

   window.Aletics = Aletics;
})();

/*
Usage:

<script id="aletics-script" src="https://<tld>/aletics/v1/tracker.js" async defer></script>
<script>
   document.querySelector("#aletics-script").onload = () => {
      (Aletics.init("http://localhost:3000/aletics")).track();
   };
</script>
*/
