let fil = document.getElementById("fil")
let bilde = document.getElementById("image")

bilde.addEventListener("change", () => {
    fil.hidden = false;
    fil.src = URL.createObjectURL(bilde.files[0]);

    fil.parentElement.querySelector(".fjern").hidden = false
})

let fjernere = document.getElementsByClassName("fjern")

function fjern(event) {
    event.target.parentElement.querySelector("#image").value = ""
    event.target.parentElement.querySelector("#fil").src = ""


    event.target.hidden = true
}