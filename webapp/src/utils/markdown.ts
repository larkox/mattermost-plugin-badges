import React from 'react';

export function markdown(text: string, siteURL?: string): React.ReactNode {
    const PostUtils = (window as any).PostUtils; // import the post utilities
    const htmlFormatedText = PostUtils.formatText(text, {siteURL});
    return PostUtils.messageHtmlToComponent(htmlFormatedText);
}
