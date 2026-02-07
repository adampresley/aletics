async function showAlert(title, message, showCancelButton, icon = null) {
   const customAlert = Swal.mixin({
      customClass: {
         confirmButton: "primary",
         cancelButton: "secondary",
      },
   });

   const result = await customAlert.fire({
      title: title,
      text: message,
      icon: icon || "info",
      showCancelButton: showCancelButton,
   });

   return result;
}

export async function showConfirm(title, message, icon = "question") {
   return await showAlert(title, message, true, icon);
}
