/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

import { DotLottie } from '@lottiefiles/dotlottie-web';
import logoUrl from './assets/lottie/wailsterm_themed.lottie?url';

function getCurrentTheme(event) {
    const newColorScheme = event.matches ? "dark" : "light";

    if (newColorScheme === 'dark') {
        return "Dark"
    } else {
        return "Light"
    }
}

const saveFirstLaunch = () => {
    localStorage.setItem('launched', 1);
}

const addClass = document.withHeader ? ' with-header' : '';
document.querySelector('#app').innerHTML += `
<div id="welcome-container" class="fade-in-image${addClass}">
    <canvas id="welcome-logo" style="width: 150px; height: 150px; margin-top: -20px; margin-bottom: -20px;"></canvas>
    <h1>Welcome to WailsTerm</h1>
    <p>Simple and lightweight terminal application</p>
    <button id="continue" style="font-size: 16px; margin-top: 10px;">Continue</button>
</div>
`;

document.querySelector('#continue').addEventListener('click', () => {
    saveFirstLaunch()
    window.location.reload();
});

const dotLottie = new DotLottie({
    autoplay: true,
    loop: false,
    canvas: document.querySelector('#welcome-logo'),
    src: logoUrl,
    themeId: getCurrentTheme(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)')),
    mode: "forward",
    speed: 1.5,
});

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (event) => {
    dotLottie.setTheme(getCurrentTheme(event));
});

document.querySelector('#welcome-logo').addEventListener('click', (e) => {
    dotLottie.setMode("reverse-bounce")
    dotLottie.setTheme(getCurrentTheme(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)')))
    dotLottie.play()
})

setInterval(async () => {
    const header = document.querySelector('#header');
    if (header && (await window.isFullscreen()) && !header.classList.contains('hidden')) {
        header.classList.add('hidden');

        const welcomeContainer = document.querySelector('#welcome-container')
        welcomeContainer.classList.remove('with-header');
        return;
    }

    if (header && !(await window.isFullscreen()) && header.classList.contains('hidden')) {
        header.classList.remove('hidden');

        const welcomeContainer = document.querySelector('#welcome-container')
        welcomeContainer.classList.add('with-header');
    }
}, 2000)