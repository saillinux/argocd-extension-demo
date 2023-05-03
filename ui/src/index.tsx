import * as React from 'react';

import {
  ActionButton,
  InfoItemRow,
  ThemeDiv,
} from 'argo-ui/v2';

import "./index.scss";

const EXTPATH = "/extensions/extdemo";

// Create a Header to access extension backend
const createHttpHeaders = () => {
  const myHeaders = new Headers();
  // The backend for this resides in the application argocd-extension-demo
  myHeaders.append("Argocd-Application-Name", "argocd:argocd-extension-demo")
  myHeaders.append("Argocd-Project-Name", "default")
  return myHeaders;
}

const getInstanceGroup = async (projectId: string, region: string, instanceGroupName: string) => {
  const myHeaders = createHttpHeaders();
  const myRequest = new Request(`${EXTPATH}/compute/instancegroup/${projectId}/${region}/${instanceGroupName}`, {
    method: "GET",
    headers: myHeaders,
    mode: "cors",
    cache: "default",
  });
  const response = await fetch(myRequest);
  const instanceGroup = await response.json();
  return instanceGroup;
};

export const Extension = (props: {
  tree: any;
  resource: any;
}) => {
  const [instanceGroup, setInstanceGroup] = React.useState(null);
  console.log(props);

  const { metadata: { annotations, name }, spec } = props.resource;
  const { location } = spec;
  
  const projectId = annotations['cnrm.cloud.google.com/project-id'];
  
  React.useEffect(() => {
    if (instanceGroup) {
      return;
    }

    (async () => {
      const instanceGroup = await getInstanceGroup(projectId, location, name);
      console.log(instanceGroup);
      setInstanceGroup(instanceGroup);
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
          label='Reset InstanceGroup'
          icon='fa-times'
          action={() => {
            setInstanceGroup(null);
          }}
        />
        <div style={{
          display: 'flex',
          alignItems: 'center',
          height: '2em'
        }}>
          Test
        </div>
      </ThemeDiv>
    </>
  );
}

export const component = Extension;