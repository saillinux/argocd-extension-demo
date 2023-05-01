import * as React from 'react';

import {
  ActionButton,
  InfoItemRow,
  ThemeDiv,
} from 'argo-ui/v2';

import "./index.scss";

export const Extension = (props: {
  tree: any;
  resource: any;
}) => {
  const [buckets, setBuckets] = React.useState([]);

  React.useEffect(() => {
    if (buckets.length > 0) {
      return;
    }
    
    (async () => {
      const myHeaders = new Headers();
      // myHeaders.append("Cookie", "")
      myHeaders.append("Argocd-Application-Name", "argocd:argocd-extension-demo")
      myHeaders.append("Argocd-Project-Name", "default")
      
      const myRequest = new Request("/extensions/extdemo/storage/list", {
        method: "GET",
        headers: myHeaders,
        mode: "cors",
        cache: "default",
      });

      const response = await fetch(myRequest);
      const listBuckets = await response.json();
      console.log(listBuckets);
      setBuckets(listBuckets);
    })();
    return () => {
      console.log("");
    };
  });

  return (
    <>
      <ThemeDiv className='info extension__info'>
        <InfoItemRow
          items={{
            content: "Instance Group",
            icon: "fa-palette",
          }}
          label='Instance Group'
        />
        <ActionButton
          label='Reset buckets'
          icon='fa-times'
          action={() => {
            setBuckets([]);
          }}
        />
        <div style={{
          display: 'flex',
          alignItems: 'center',
          height: '2em'
        }}>
          {buckets.map((bucket: any) => {
            return (
              <InfoItemRow
                items={{
                  content: bucket,
                  icon: 'fa-shoe-prints',
                }}
                label={bucket}
              />
            )
          })}
        </div>
      </ThemeDiv>
    </>
  );
}

export const component = Extension;