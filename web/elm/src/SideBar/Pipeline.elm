module SideBar.Pipeline exposing (pipeline)

import Assets
import Concourse
import HoverState
import Message.Message exposing (DomID(..), Message(..), PipelinesSection(..))
import Routes
import Set exposing (Set)
import SideBar.Styles as Styles
import SideBar.Views as Views


type alias PipelineScoped a =
    { a
        | teamName : String
        , pipelineName : String
    }


pipeline :
    { a
        | hovered : HoverState.HoverState
        , currentPipeline : Maybe (PipelineScoped b)
        , favoritedPipelines : Set Int
        , isFavoritesSection : Bool
    }
    -> Concourse.Pipeline
    -> Views.Pipeline
pipeline params p =
    let
        isCurrent =
            case params.currentPipeline of
                Just cp ->
                    cp.pipelineName == p.name && cp.teamName == p.teamName

                Nothing ->
                    False

        pipelineId =
            Concourse.toPipelineId p

        domID =
            SideBarPipeline
                (if params.isFavoritesSection then
                    FavoritesSection

                 else
                    AllPipelinesSection
                )
                pipelineId

        isHovered =
            HoverState.isHovered domID params.hovered

        isFavorited =
            Set.member p.id params.favoritedPipelines
    in
    { icon =
        if p.archived then
            Assets.ArchivedPipelineIcon

        else if isHovered then
            Assets.PipelineIconWhite

        else if isCurrent then
            Assets.PipelineIconLightGrey

        else
            Assets.PipelineIconGrey
    , name =
        { color =
            if isHovered then
                Styles.White

            else if isCurrent then
                Styles.LightGrey

            else
                Styles.Grey
        , text = p.name
        , weight =
            if isCurrent || isHovered then
                Styles.Bold

            else
                Styles.Default
        }
    , background =
        if isCurrent then
            Styles.Dark

        else if isHovered then
            Styles.Light

        else
            Styles.Invisible
    , href =
        Routes.toString <|
            Routes.Pipeline { id = pipelineId, groups = [] }
    , domID = domID
    , starIcon =
        { filled = isFavorited
        , isBright = isHovered || isCurrent
        }
    , id = pipelineId
    , databaseID = p.id
    }
