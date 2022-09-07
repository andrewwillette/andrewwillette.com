import {Component} from 'react';
import "./homePage.css";

export class HomePage extends Component<any, any> {
	readonly homeImageUrl = "https://andrewwillette.s3.us-east-2.amazonaws.com/newdir/website_photo_2.jpeg";
    render() {
        return (
            <>
                <img src={this.homeImageUrl} className="personalImage" alt="logo" />
                <div id="home-page">
                    <p>
                        Hi! My name is Andrew Willette. I am a software developer based in Madison, Wisconsin.
                    </p>
                    <p>
                        I like playing violin and host some recordings on my site here.
                    </p>
                </div>
            </>
        );
    }
}
