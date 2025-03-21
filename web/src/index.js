document.getElementById('audio').addEventListener('click', () => {
    // alert('Button clicked!');
    console.log("button is clicked");

    const fileName = "beef\\Episode 1 - Dr David Pin.mp3";
    const url = `http://localhost:8080/episode?name=${encodeURIComponent(fileName)}`;
    console.log(url);

    fetch(url)
        .then(response => response.blob())
        .then(blob => {
            const url = URL.createObjectURL(blob);
            const audio = new Audio(url);
            audio.play();
        });
});


document.getElementById('podcast').addEventListener('click', () => {
    console.log("podcast button is clicked");

    const fileName = "beef";
    const url = `http://localhost:8080/podcast?title=${encodeURIComponent(fileName)}`;
    console.log(url);

    fetch(url)
        .then(response => response.json())
        .then(data => {
            console.log("Podcast data:", data);
        })
        .catch(error => {
            console.error("Error fetching podcast data:", error);
        });
})