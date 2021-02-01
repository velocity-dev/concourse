module Message.Message exposing
    ( DomID(..)
    , DropTarget(..)
    , Message(..)
    , PipelinesSection(..)
    , VersionId
    , VersionToggleAction(..)
    , VisibilityAction(..)
    )

import Concourse
import Concourse.Cli as Cli
import Concourse.Pagination exposing (Page)
import Routes exposing (StepID)
import StrictEvents


type Message
    = -- Top Bar
      FilterMsg String
    | FocusMsg
    | BlurMsg
      -- Pipeline
    | ToggleGroup Concourse.PipelineGroup
    | SetGroups (List String)
      -- Dashboard
    | DragStart String Int
    | DragOver DropTarget
    | DragEnd
      -- Resource
    | EditComment String
    | FocusTextArea
    | BlurTextArea
      -- Build
    | ScrollBuilds StrictEvents.WheelEvent
    | RevealCurrentBuildInHistory
    | SetHighlight String Int
    | ExtendHighlight String Int
      -- common
    | Hover (Maybe DomID)
    | Click DomID
    | GoToRoute Routes.Route
    | Scrolled StrictEvents.ScrollState


type DomID
    = ToggleJobButton
    | TriggerBuildButton
    | AbortBuildButton
    | RerunBuildButton
    | PreviousPageButton
    | NextPageButton
    | CheckButton Bool
    | EditButton
    | SaveCommentButton
    | ResourceCommentTextarea
    | ChangedStepLabel StepID String
    | StepState StepID
    | PinIcon
    | PinMenuDropDown String
    | PinButton VersionId
    | PinBar
    | PipelineCardName PipelinesSection Concourse.DatabaseID
    | InstanceGroupCardName PipelinesSection Concourse.TeamName String
    | PipelineCardNameHD Concourse.DatabaseID
    | InstanceGroupCardNameHD Concourse.TeamName String
    | PipelineCardInstanceVar PipelinesSection Concourse.DatabaseID String String
    | PipelineStatusIcon PipelinesSection Concourse.DatabaseID
    | PipelineCardPauseToggle PipelinesSection Concourse.PipelineIdentifier
    | TopBarFavoritedIcon Concourse.DatabaseID
    | TopBarPauseToggle Concourse.PipelineIdentifier
    | VisibilityButton PipelinesSection Concourse.DatabaseID
    | PipelineCardFavoritedIcon PipelinesSection Concourse.DatabaseID
    | FooterCliIcon Cli.Cli
    | WelcomeCardCliIcon Cli.Cli
    | CopyTokenButton
    | SendTokenButton
    | CopyTokenInput
    | JobGroup Int
    | StepTab String Int
    | StepHeader String
    | StepSubHeader String Int
    | StepInitialization String
    | ShowSearchButton
    | ClearSearchButton
    | LoginButton
    | LogoutButton
    | UserMenu
    | PaginationButton Page
    | VersionHeader VersionId
    | VersionToggle VersionId
    | BuildTab Int String
    | JobPreview PipelinesSection Concourse.DatabaseID Concourse.JobName
    | PipelinePreview PipelinesSection Concourse.DatabaseID
    | HamburgerMenu
    | SideBarResizeHandle
    | SideBarTeam PipelinesSection String
    | SideBarPipeline PipelinesSection Concourse.PipelineIdentifier
    | SideBarInstanceGroup PipelinesSection Concourse.TeamName String
    | SideBarFavoritedIcon Concourse.DatabaseID
    | Dashboard
    | DashboardGroup String


type PipelinesSection
    = FavoritesSection
    | AllPipelinesSection


type VersionToggleAction
    = Enable
    | Disable


type VisibilityAction
    = Expose
    | Hide


type alias VersionId =
    Concourse.VersionedResourceIdentifier


type DropTarget
    = Before Int
    | End
