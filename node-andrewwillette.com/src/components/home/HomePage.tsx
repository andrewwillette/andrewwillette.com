import {Component} from 'react';
import homepage_photo from './website_photo_2.jpeg';
import "./homePage.css";

export class HomePage extends Component<any, any> {
    render() {
        return (
            <>
                <img src={homepage_photo} className="personalImage" alt="logo" />
                <div id="home-page">
                    <p>
                        Hi! My name is Andrew Willette. I am a software developer based in Kansas City, Kansas. 
                    </p>
                    <p>
                        I like playing violin and host some recordings on my site here.
                    </p>
                </div>
            </>
        );
    }
}
