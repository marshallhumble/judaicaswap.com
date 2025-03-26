document.addEventListener("DOMContentLoaded", function () {
    const form = document.getElementById("createform");
    const s3ImageURL = "https://judaicaswap.s3.us-east-1.amazonaws.com/images/"
    let imageTag0 = document.getElementById("picture0")
    let imageTag1 = document.getElementById("picture1")
    let imageTag2 = document.getElementById("picture2")
    let imageTag3 = document.getElementById("picture3")
    let imageTag4 = document.getElementById("picture4")

    form.addEventListener('submit', async (event) => {
        //We need to do some things before we submit
        event.preventDefault();

        //Get all the files and iterate over them
        const files = document.querySelector('input[type="file"]').files;

        for (let i = 0; i < files.length; i++) {

            //5 is the max (start count from 0)
            if (i > 4) {
                break
            }

            //set the working file
            const file = files[i];

            if (file === 'null') {
                break
            }

            //create a unique uuid and add back the file extension
            const uuidName = crypto.randomUUID().replace(/-/g, '');
            const ext = file.name.slice((file.name.lastIndexOf(".") - 1 >>> 0) + 2);

            //Get the image element, and rename the file
            let fileName = uuidName + "." + ext

            //get the pre-signed URL using the new filename
            const s3URL = await fetch(`/api/v1/signed/${ext}/${fileName}`)
                .then(response => response.json()).catch(error => console.log(error));

            // post the image directly to the s3 bucket with a PUT request
            const response = await fetch(s3URL, {
                method: "PUT",
                headers: {
                    "Content-Type": `image/${ext}`
                },
                body: file
            })
            
            //add the URL to the file so we can just use that directly in forms
            fileName = s3ImageURL + fileName

            switch (i) {
                case 0:
                    imageTag0.value = fileName
                    break
                case 1:
                    imageTag1.value = fileName
                    break
                case 2:
                    imageTag2.value = fileName
                    break
                case 3:
                    imageTag3.value = fileName
                    break
                case 4:
                    imageTag4.value = fileName
                    break
            }
        }

        form.submit();
    });
});