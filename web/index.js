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