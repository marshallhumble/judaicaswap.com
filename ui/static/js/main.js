document.addEventListener("DOMContentLoaded", function () {
  const signUpButton = document.getElementById("signUp");
  if (signUpButton != null) {
    signUpButton.addEventListener("click", function () {
      window.location.href = "/user/signup";
    });
  }

  const loginButton = document.getElementById("logIn");
  if (loginButton != null){
  loginButton.addEventListener("click", function () {
    window.location.href = "/user/login";
  })
  }

  const logoutButton = document.getElementById("logOut");
  if (logoutButton != null) {
    logoutButton.addEventListener("click", function () {
    window.location.href = "/user/logout";
  })
  }

});
