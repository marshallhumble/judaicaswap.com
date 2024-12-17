document.addEventListener("DOMContentLoaded", function () {
  const signUpButton = document.getElementById("signUp");
  signUpButton.addEventListener("click", function () {
    window.location.href = "/user/signup";
  });
  const loginButton = document.getElementById("logIn");
  loginButton.addEventListener("click", function () {
    window.location.href = "/user/login";
  });
  const logoutButton = document.getElementById("logOut");
  loginButton.addEventListener("click", function () {
    window.location.href = "/user/logout";
  });
});
