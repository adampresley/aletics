import { showConfirm } from "/static/js/components/alert.js";

/*
Catch all confirms and display a confirmation dialog before executing the request.
*/
document.addEventListener("htmx:confirm", async (e) => {
   if (!e.detail.question) {
      return;
   }

   e.preventDefault();

   const response = await showConfirm("Are you sure?", e.detail.question, "warning");

   if (response.isConfirmed) {
      e.detail.issueRequest(true);
   }
});
