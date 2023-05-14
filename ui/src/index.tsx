import * as React from 'react';
import {
  ActionButton,
  EffectDiv,
  InfoItemKind,
  InfoItemRow,
  // Menu,
  ThemeDiv,
  Tooltip,  
  WaitFor,
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
  const myRequest = new Request(`${EXTPATH}/compute/instancegroup/get/${projectId}/${region}/${instanceGroupName}`, {
    method: "GET",
    headers: myHeaders,
    mode: "cors",
    cache: "default",
  });
  const response = await fetch(myRequest);
  const instanceGroup = await response.json();
  return instanceGroup;
};

const listInstanceTemplates = async (projectId: string) => {
  const myHeaders = createHttpHeaders();
  const myRequest = new Request(`${EXTPATH}/compute/instancetemplate/list/${projectId}`, {
    method: "GET",
    headers: myHeaders,
    mode: "cors",
    cache: "default",
  });
  const response = await fetch(myRequest);
  const instanceTemplates = await response.json();
  return instanceTemplates;
};

// /compute/instancegroup/update/heewonk-bunker/us-west1/heewonk-bunker-argocd-cc-demo-instance-group?strategy=canary&target_template=heewonk-bunker-argocd-cc-demo-instance-template-12
const deployRevision = async (projectId: string, region: string, instanceGroupName: string, startegy: string, instanceTemplate: string) => {
  const myHeaders = createHttpHeaders();
  const params = {
    strategy: startegy,
    target_template: instanceTemplate,
  };
  const queryParams = new URLSearchParams(params).toString();
  const myRequest = new Request(`${EXTPATH}/compute/instancegroup/update/${projectId}/${region}/${instanceGroupName}?` + queryParams, {
    method: "GET",
    headers: myHeaders,
    mode: "cors",
    cache: "default",
  });  
  const response = await fetch(myRequest);
  const deployStatus = await response.json();
  return deployStatus;
};

const RolloutSummary = (
  props: {
    resource: any;
  }
) => {    
  const { spec: { location, targetSize, updatePolicy, distributionPolicy } } = props.resource;

  return (
    <ThemeDiv className='info rollout__info'>    
      <div className='info__title'>
        Summary
      </div>
      <ThemeDiv className='rollout__info__section'>
        <InfoItemRow
          items={{
            content: location,
            kind: InfoItemKind.BlueGreen,
            icon: 'fa-map-marker'
          }}
          label='Location'
        />
        <InfoItemRow
          items={{
            kind: InfoItemKind.Canary,
            content: targetSize,
            icon: 'fa-balance-scale-right'
          }}
          label='Target Size'
        />
        <InfoItemRow
          items={{
            kind: InfoItemKind.Colored,
            content: distributionPolicy.targetShape,
            icon: 'fa-balance-scale'
          }}
          label='Distribution Policy'
        />
        <InfoItemRow
          items={{
            kind: InfoItemKind.Colored,
            content: updatePolicy.instanceRedistributionType,
          }}
          label='Update Policy'
        />
        <InfoItemRow
          items={{
            content: updatePolicy.maxSurge.fixed,
          }}
          label='Max Surge'
        />
        <InfoItemRow
          items={{
            content: updatePolicy.maxUnavailable.fixed,
          }}
          label='Max Unavailable'
        />                        
        {' '}
      </ThemeDiv>
    </ThemeDiv>
  );
};

const RolloutStatus = (
  props: {
    application: any;
    instanceGroup: any;
  }  
) => {    
  const { application, instanceGroup } = props;  
  const { spec: { source: { helm: { parameters } } } } = application;

  return (
    <ThemeDiv className='info rollout__info'>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        height: '2em'
      }}>
        <ThemeDiv className='info__title' style={{marginBottom: '0'}}>
            Status
        </ThemeDiv>
      </div>
      <div style={{
        margin: '1em 0',
        whiteSpace: 'nowrap'
      }}>
        <InfoItemRow
          items={{
            kind: instanceGroup ? instanceGroup.status ?  InfoItemKind.BlueGreen : InfoItemKind.Colored : InfoItemKind.Default,
            content: instanceGroup ? instanceGroup.status ? 'Stable' : 'Updating' : 'Unknown',
          }}
          label='Current Status'
        />
        <InfoItemRow
          items={{
            kind: instanceGroup ? instanceGroup.status ?  InfoItemKind.BlueGreen : InfoItemKind.Colored : InfoItemKind.Default,
            content: instanceGroup ? instanceGroup.status ? "None" : "Updating" : 'Unknown',                
          }}
          label='Current Action'              
        />
        <InfoItemRow
          items={{
            content: instanceGroup ? instanceGroup.managedInstances.length.toString() : "0",
          }}
          label='Current Size'
        />
        <InfoItemRow
          items={{
            content: parameters[1]['value'],
          }}
          label={parameters[1]['name']}
        />               
        <InfoItemRow
          items={{
            content: parameters[0]['value'],
          }}
          label={parameters[0]['name']}              
        />
      </div>
    </ThemeDiv>    
  )
};

const RolloutTop = (
  props: {
    application: any;
    resource: any;
    instanceGroup: any;
  }  
) => {
  const { application, resource, instanceGroup } = props;  
  return (
    <ThemeDiv
      className='rollout__row rollout__row--top'
    >
      <RolloutSummary
        resource={resource}
      />
      <RolloutStatus
        application={application}
        instanceGroup={instanceGroup}     
      />
    </ThemeDiv>
  );
};

const RolloutRevision = (
  props: {
    revision: string;
    instanceGroup: any;
  }
) => {
  const { revision, instanceGroup } = props;

  if (!instanceGroup) {
    return null;
  }
  
  const { projectId, region, groupName } = instanceGroup;
  const [collapsed, setCollapsed] = React.useState(true);

  return (
    <EffectDiv
      key={revision}
      className='revision'
    >
      <ThemeDiv className='revision__header'>
        {revision}
        <div style={{marginLeft: 'auto', display: 'flex', alignItems: 'center'}}>
          <ActionButton
              action={() => {                
                (async () => {
                  const status = await deployRevision(projectId, region, groupName, "rollout", revision);
                  console.log("Status: ", status);
                })();
              }}
              label='DEPLOY'
              icon='fa-undo-alt'
              style={{fontSize: '12px'}}
          />
          <ThemeDiv className='revision__header__button' onClick={() => setCollapsed(!collapsed)}>
            <i className={`fa ${collapsed ? 'fa-chevron-circle-down' : 'fa-chevron-circle-up'}`} />
          </ThemeDiv>
        </div>
      </ThemeDiv>
      <ThemeDiv className='revision__images'>
        <div key={"rsInfo.objectMeta.uid"} style={{marginBottom: '1em'}}>
          <ThemeDiv className='pods'>
            <ThemeDiv className='pods__header'>
              <div style={{marginRight: 'auto', flexShrink: 0}}>
                {instanceGroup && instanceGroup.instanceGroupTemplate}
              </div>
            </ThemeDiv>
            <ThemeDiv className='pods__container'>
              <WaitFor loading={(instanceGroup && instanceGroup.managedInstances || []).length < 1}>
                {
                  instanceGroup && instanceGroup.managedInstances.map((instance: any) => {
                    return (
                      <Tooltip content={
                        <div>
                          <div>Name: {instance.instance}</div>
                          <div>Status: {instance.status}</div>
                          <div>Zone: {instance.zone}</div>
                        </div>
                      }>
                        <ThemeDiv className={`pod-icon pod-icon--success`}>
                          <i className={`fa ${instance.status === 'RUNNING' ? 'fa-check-circle' : 'fa-exclamation-triangle'}`} />
                        </ThemeDiv>
                      </Tooltip>
                    );
                  })
                }
              </WaitFor>                      
            </ThemeDiv>
          </ThemeDiv>
        </div>
      </ThemeDiv>
    </EffectDiv>
  );
};

const RolloutRevisions = (
  props: {
    application: any;
    projectId: string;
    instanceGroup: any;
  }
) => {
  const { application: { metadata: { name } }, projectId, instanceGroup } = props;
  const [revisions, setRevisions] = React.useState([]);
  // console.log("application name: ", name);

  React.useEffect(() => {
    if (revisions.length > 0) {
      return;
    }

    (async () => {
      const instanceTemplates = await listInstanceTemplates(projectId);
      const regexMatch = new RegExp(name);
      const instanceTemplatesFiltered = instanceTemplates.filter((instanceTemplate: string) => regexMatch.test(instanceTemplate));
      instanceTemplatesFiltered.sort().reverse();
      // console.log("filtered: ", instanceTemplatesFiltered);
      setRevisions(instanceTemplatesFiltered);
    })();
    return () => {
      console.log("");
    };
  });

  return (
    <ThemeDiv className='info rollout__info rollout__revisions'>
      <div className='info__title'>
        Revisions
      </div>
      <div style={{ marginTop: '1em' }}>
        {
          revisions.map((revision: any, index: number) => {
            return (
              <RolloutRevision
                key={`revision-${index}`}
                revision={revision}
                instanceGroup={instanceGroup}
              />
            );
          })
        }       
      </div>
    </ThemeDiv>
  );
};

const RolloutHistory = (
  props: {
    application: any;
  }
) => {
  const { status: { history }} = props.application;

  return (
    <ThemeDiv className='info steps'>
      <ThemeDiv className='info__title'>History</ThemeDiv>
      <div style={{marginTop: '1em'}}>
        {
          history.map((item: any, index: number) => {
            const { deployedAt, source: { helm: { parameters } } } = item;
            const [image, version] = parameters;

            return (
              <EffectDiv
                key={`history-${index}`}
                className='revision'
              >
                <ThemeDiv className='revision__header'>
                  {deployedAt.split('T')[0]}
                </ThemeDiv>
                <InfoItemRow
                  items={{
                    content: version.value,
                  }}
                  label={version.name}
                />
                <InfoItemRow
                  items={{
                    content: image.value,
                  }}
                  label={image.name}
                />                    
              </EffectDiv>
            );
          })
        }
      </div>
    </ThemeDiv>
  );
};

const RolloutBottom = (
  props: {
    application: any;
    projectId: string;
    instanceGroup: any;
  }
) => {
  const { application, projectId, instanceGroup } = props;

  return (
    <ThemeDiv className='rollout__row rollout__row--bottom'>
      <RolloutRevisions
        application={application}
        projectId={projectId}
        instanceGroup={instanceGroup}
      />
      <RolloutHistory
        application={application}
      />   
    </ThemeDiv>
  );
};

export const Extension = (props: {
  application: any;
  tree: any;
  resource: any;
}) => {
  console.log(props);
  const [instanceGroup, setInstanceGroup] = React.useState(null);
  const { metadata: { annotations, name }, spec: { location } } = props.resource;

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
      <RolloutTop
        application={props.application}
        resource={props.resource}
        instanceGroup={instanceGroup}
      />
      <RolloutBottom
        application={props.application}
        projectId={projectId}
        instanceGroup={instanceGroup}
      />      
    </>
  );
};

export const component = Extension;