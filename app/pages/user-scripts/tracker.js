(() => {
   if (window.Aletics) {
      return;
   }

   function getSessionId() {
      let sid = sessionStorage.getItem('aletics_sid');
      if (!sid) {
         sid = ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
            (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
         );
         sessionStorage.setItem('aletics_sid', sid);
      }
      return sid;
   }

   function assemble(token) {
      return {
         token: token,
         path: window.location.pathname,
         queryString: window.location.search,
         browser: getBrowserName(),
         sessionId: getSessionId(),
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

   function newTracker(baseUrl, token) {
      const t = {
         track: () => {
            const payload = assemble(token);
            const data = JSON.stringify(payload);
            const endpoint = `${baseUrl}/v1/track`.replace(/([^:]\/)\/+/g, "$1");

            if (navigator.sendBeacon) {
               navigator.sendBeacon(endpoint, data);
            } else {
               fetch(endpoint, {
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
      init: (baseUrl, token) => {
         return newTracker(baseUrl, token);
      },
   };

   window.Aletics = Aletics;
})();

/*
Usage:

<script id="aletics-script" src="https://<tld>/aletics/v1/tracker.js" async defer></script>
<script>
   document.querySelector("#aletics-script").onload = () => {
      (Aletics.init("https://<tld>/aletics", "<property token>")).track();
   };
</script>

*/
